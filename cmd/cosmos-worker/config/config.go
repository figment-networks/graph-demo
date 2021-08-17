package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	Name      = "cosmos-worker"
	Version   string
	GitSHA    string
	Timestamp string
)

// Config holds the configuration data
type Config struct {
	AppEnv string `json:"app_env" envconfig:"APP_ENV" default:"development"`

	CosmosGRPCAddr string `json:"cosmos_grpc_addr" envconfig:"COSMOS_GRPC_ADDR"`
	ChainID        string `json:"chain_id" envconfig:"CHAIN_ID"`

	ManagerURL string `json:"managers" envconfig:"MANAGER_URL" default:"ws://0.0.0.0:8085"`

	TimeoutBlockCall       time.Duration `json:"timeout_block_call" envconfig:"TIMEOUT_BLOCK_CALL" default:"30s"`
	TimeoutTransactionCall time.Duration `json:"timeout_transaction_call" envconfig:"TIMEOUT_TRANSACTION_CALL" default:"30s"`
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
