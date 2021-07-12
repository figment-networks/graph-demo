package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/cosmos-worker/config"
	"github.com/figment-networks/graph-demo/cosmos-worker/api"
	"github.com/figment-networks/graph-demo/cosmos-worker/client"
	"github.com/figment-networks/indexer-manager/worker/connectivity"
	grpcIndexer "github.com/figment-networks/indexer-manager/worker/transport/grpc"
	grpcProtoIndexer "github.com/figment-networks/indexer-manager/worker/transport/grpc/indexer"

	"github.com/figment-networks/indexing-engine/health"
	"github.com/figment-networks/indexing-engine/metrics"
	"github.com/figment-networks/indexing-engine/metrics/prometheusmetrics"

	"github.com/google/uuid"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type flags struct {
	configPath  string
	showVersion bool
}

var configFlags = flags{}

func init() {
	flag.BoolVar(&configFlags.showVersion, "v", false, "Show application version")
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// Initialize configuration
	cfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatalf("error initializing config [ERR: %v]", err.Error())
	}

	if cfg.RollbarServerRoot == "" {
		cfg.RollbarServerRoot = "github.com/figment-networks/graph-demo/cosmos-worker"
	}
	rcfg := &logger.RollbarConfig{
		AppEnv:             cfg.AppEnv,
		RollbarAccessToken: cfg.RollbarAccessToken,
		RollbarServerRoot:  cfg.RollbarServerRoot,
		Version:            config.GitSHA,
		ChainIDs:           []string{cfg.ChainID},
	}

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		logger.Init("console", "debug", []string{"stderr"}, rcfg)
	} else {
		logger.Init("json", "info", []string{"stderr"}, rcfg)
	}

	logger.Info(config.IdentityString())
	defer logger.Sync()

	// Initialize metrics
	prom := prometheusmetrics.New()
	err = metrics.AddEngine(prom)
	if err != nil {
		logger.Error(err)
	}
	err = metrics.Hotload(prom.Name())
	if err != nil {
		logger.Error(err)
	}

	workerRunID, err := uuid.NewRandom() // UUID V4
	if err != nil {
		logger.Error(fmt.Errorf("error generating UUID: %w", err))
		return
	}

	managers := strings.Split(cfg.Managers, ",")
	hostname := cfg.Hostname
	if hostname == "" {
		hostname = cfg.Address
	}

	logger.Info(fmt.Sprintf("Self-hostname (%s) is %s:%s ", workerRunID.String(), hostname, cfg.Port))

	c := connectivity.NewWorkerConnections(workerRunID.String(), hostname+":"+cfg.Port, "cosmos", cfg.ChainID, "0.0.1")
	for _, m := range managers {
		c.AddManager(m + "/client_ping")
	}

	logger.Info(fmt.Sprintf("Connecting to managers (%s)", strings.Join(managers, ",")))

	go c.Run(ctx, logger.GetLogger(), cfg.ManagerInterval)

	if cfg.CosmosGRPCAddr == "" {
		logger.Error(fmt.Errorf("cosmos grpc address is not set"))
		return
	}
	grpcConn, dialErr := grpc.DialContext(ctx, cfg.CosmosGRPCAddr, grpc.WithInsecure())
	if dialErr != nil {
		logger.Error(fmt.Errorf("error dialing grpc: %w", dialErr))
		return
	}
	defer grpcConn.Close()

	apiClient := api.NewClient(logger.GetLogger(), grpcConn, &api.ClientConfig{
		ReqPerSecond:        int(cfg.RequestsPerSecond),
		TimeoutBlockCall:    cfg.TimeoutBlockCall,
		TimeoutSearchTxCall: cfg.TimeoutTransactionCall,
	})

	grpcServer := grpc.NewServer()
	workerClient := client.NewIndexerClient(ctx, logger.GetLogger(), apiClient, uint64(cfg.MaximumHeightsToGet))

	worker := grpcIndexer.NewIndexerServer(ctx, workerClient, logger.GetLogger())
	grpcProtoIndexer.RegisterIndexerServiceServer(grpcServer, worker)

	mux := http.NewServeMux()

	monitor := &health.Monitor{}
	go monitor.RunChecks(ctx, cfg.HealthCheckInterval)
	monitor.AttachHttp(mux)

	attachDynamic(ctx, mux)

	s := &http.Server{
		Addr:         "0.0.0.0:" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go runGRPC(grpcServer, cfg.Port, logger.GetLogger(), exit)
	go runHTTP(s, cfg.HTTPPort, logger.GetLogger(), exit)

RunLoop:
	for {
		select {
		case sig := <-osSig:
			logger.Info("Stopping worker... ", zap.String("signal", sig.String()))
			cancel()
			logger.Info("Canceled context, gracefully stopping grpc")
			grpcServer.Stop()
			logger.Info("Stopped grpc, stopping http")
			err := s.Shutdown(ctx)
			if err != nil {
				logger.GetLogger().Error("Error stopping http server ", zap.Error(err))
			}
			break RunLoop
		case k := <-exit:
			logger.Info("Stopping worker... ", zap.String("reason", k))
			cancel()
			logger.Info("Canceled context, gracefully stopping grpc")
			if k == "grpc" { // (lukanus): when grpc is finished, stop http and vice versa
				err := s.Shutdown(ctx)
				if err != nil {
					logger.GetLogger().Error("Error stopping http server ", zap.Error(err))
				}
			} else {
				grpcServer.Stop()
			}
			break RunLoop
		}
	}

}

func initConfig(path string) (*config.Config, error) {
	cfg := &config.Config{}
	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return nil, err
		}
	}

	if cfg.CosmosGRPCAddr != "" {
		return cfg, nil
	}

	if err := config.FromEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func runGRPC(grpcServer *grpc.Server, port string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[GRPC] Listening on 0.0.0.0:%s", port))
	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		logger.Error("[GRPC] failed to listen", zap.Error(err))
		exit <- "grpc"
		return
	}

	// (lukanus): blocking call on grpc server
	grpcServer.Serve(lis)
	exit <- "grpc"
}

func runHTTP(s *http.Server, port string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on 0.0.0.0:%s", port))
	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
}
