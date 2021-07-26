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
	apiTransportWS "github.com/figment-networks/graph-demo/cosmos-worker/api/transport/ws"
	"github.com/figment-networks/graph-demo/cosmos-worker/client"

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

	log := logger.GetLogger()

	if cfg.CosmosGRPCAddr == "" {
		log.Error("cosmos grpc address is not set")
		return
	}
	grpcConn, dialErr := grpc.DialContext(ctx, cfg.CosmosGRPCAddr, grpc.WithInsecure())
	if dialErr != nil {
		log.Error("error dialing grpc: %w", zap.Error(dialErr))
		return
	}
	defer grpcConn.Close()

	cliCfg := &client.ClientConfig{
		TimeoutBlockCall:    cfg.TimeoutBlockCall,
		TimeoutSearchTxCall: cfg.TimeoutTransactionCall,
	}

	apiClient := client.NewClient(logger.GetLogger(), grpcConn, cliCfg, "mainnet")
	wstr := apiTransportWS.NewProcessHandler(logger.GetLogger(), apiClient)
	apiClient.LinkPersistor(wstr)

	if err := wstr.Connect(ctx, cfg.ManagerURL); err != nil {
		log.Error("error connecting to manager ", zap.Error(err), zap.String("address", cfg.ManagerURL))
		return
	}

	if err := wstr.Register(ctx, cfg.ChainID); err != nil {
		log.Error("error registering  to manager ", zap.Error(err), zap.String("chainID", cfg.ChainID))
		return
	}

	mux := http.NewServeMux()
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
