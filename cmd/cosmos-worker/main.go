package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/cosmos-worker/config"
	transportHTTP "github.com/figment-networks/graph-demo/cosmos-worker/api/transport/http"
	"github.com/figment-networks/graph-demo/cosmos-worker/client"

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

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		logger.Init("console", "debug", []string{"stderr"})
	} else {
		logger.Init("json", "info", []string{"stderr"})
	}

	logger.Info(config.IdentityString())
	defer logger.Sync()

	workerRunID, err := uuid.NewRandom() // UUID V4
	if err != nil {
		logger.Error(fmt.Errorf("error generating UUID: %w", err))
		return
	}

	hostname := cfg.Hostname
	if hostname == "" {
		hostname = cfg.Address
	}

	logger.Info(fmt.Sprintf("Self-hostname (%s) is %s:%s ", workerRunID.String(), hostname, cfg.Port))

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

	cliCfg := &client.ClientConfig{
		ReqPerSecond:        int(cfg.RequestsPerSecond),
		TimeoutBlockCall:    cfg.TimeoutBlockCall,
		TimeoutSearchTxCall: cfg.TimeoutTransactionCall,
	}

	apiClient := client.New(logger.GetLogger(), grpcConn, cliCfg, "mainnet")

	mux := http.NewServeMux()

	httpHandler := transportHTTP.NewHandler(apiClient)
	httpHandler.AttachToMux(mux)

	s := &http.Server{
		Addr:         "0.0.0.0:" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go runHTTP(s, cfg.Port, logger.GetLogger(), exit)

RunLoop:
	for {
		select {
		case sig := <-osSig:
			logger.Info("Stopping worker... ", zap.String("signal", sig.String()))
			cancel()
			logger.Info("Canceled context, gracefully stopping grpc")
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

func runHTTP(s *http.Server, port string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on 0.0.0.0:%s", port))
	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
}
