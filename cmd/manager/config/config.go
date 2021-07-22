package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	Name      = "indexer-manager"
	Version   string
	GitSHA    string
	Timestamp string
)

const (
	modeDevelopment = "development"
	modeProduction  = "production"
)

// Config holds the configuration data
type Config struct {
	AppEnv  string `json:"app_env" envconfig:"APP_ENV" default:"development"`
	Address string `json:"address" envconfig:"ADDRESS" default:"127.0.0.1:8085"`

	StartHeight      uint64 `json:"start_height" envconfig:"START_HEIGHT" default:"6999263"`
	DatabaseURL      string `json:"database_url" envconfig:"DATABASE_URL"`
	WorkerConfigPath string `json:"worker_config_path" envconfig:"WORKER_CONFIG_PATH"`

	GrpcMaxRecvSize int `json:"grpc_max_recv_size" envconfig:"GRPC_MAX_RECV_SIZE" default:"1073741824"` // 1024^3
	GrpcMaxSendSize int `json:"grpc_max_send_size" envconfig:"GRPC_MAX_SEND_SIZE" default:"1073741824"` // 1024^3

	HealthCheckInterval time.Duration `json:"health_check_interval" envconfig:"HEALTH_CHECK_INTERVAL" default:"10s"`
}

type WorkerConfig struct {
	WorkerAddrs []WorkerAddr `json:"worker_addrs"`
}

type WorkerAddr struct {
	ChainID string `json:"chain_id"`
	URL     string `json:"url"`
}

// FromFile reads the config from a file
func FromFile(path string, config *Config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

// FromEnv reads the config from environment variables
func FromEnv(config *Config) error {
	return envconfig.Process("", config)
}
