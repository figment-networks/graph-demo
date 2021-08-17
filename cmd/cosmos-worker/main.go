package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/cosmos-worker/config"
	apiTransportWS "github.com/figment-networks/graph-demo/cosmos-worker/api/transport/ws"
	"github.com/figment-networks/graph-demo/cosmos-worker/client"

	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type flags struct {
	configPath string
}

var configFlags = flags{}

func init() {
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.Parse()
}

func main() {
	ctx := context.Background()
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

	apiClient := client.NewClient(logger.GetLogger(), grpcConn, cliCfg)
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

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	sig := <-osSig
	logger.Info("Stopping worker... ", zap.String("signal", sig.String()))
	logger.Info("Canceled context, gracefully stopping grpc")
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
