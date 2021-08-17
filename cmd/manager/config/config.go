package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

var (
	Name      = "manager"
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
	AppEnv        string `json:"app_env" envconfig:"APP_ENV" default:"development"`
	Address       string `json:"address" envconfig:"ADDRESS" default:"127.0.0.1:8085"`
	DatabaseURL   string `json:"database_url" envconfig:"DATABASE_URL"`
	LowestHeights string `json:"lowest_heights" envconfig:"LOWEST_HEIGHTS"`
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
