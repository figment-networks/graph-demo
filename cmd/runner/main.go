package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/cmd/runner/config"
	"github.com/figment-networks/graph-demo/runner/api/httpapi"
	"go.uber.org/zap"
)

type flags struct {
	configPath string
}

var configFlags = flags{}

func init() {
	flag.StringVar(&configFlags.configPath, "config", "", "path to config.json file")
	flag.Parse()
}

func main() {
	cfg, err := getConfig(configFlags.configPath)
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

	mux := http.NewServeMux()

	httpapi.AttachMux(mux)

	s := &http.Server{
		Addr:         cfg.Address + ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 40 * time.Second,
	}

	logger.Info("[HTTP] Listening on", zap.String("address", cfg.Address), zap.String("port", cfg.HTTPPort))
	if err := s.ListenAndServe(); err != nil {
		logger.GetLogger().Error("[GRPC] Error while listening ", zap.String("address", cfg.Address), zap.String("port", cfg.HTTPPort), zap.Error(err))
	}

}

func getConfig(path string) (cfg *config.Config, err error) {
	cfg = &config.Config{}
	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return nil, err
		}
	}

	if err := config.FromEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
