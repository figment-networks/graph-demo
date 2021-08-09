package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

var (
	Name      = "runner"
	Version   string
	GitSHA    string
	Timestamp string
)

type Config struct {
	AppEnv string `json:"app_env" envconfig:"APP_ENV" default:"development"`

	Address  string `json:"address" envconfig:"ADDRESS" default:"0.0.0.0"`
	HTTPPort string `json:"http_port" envconfig:"HTTP_PORT" default:"8098"`

	// A comma separated list of paths to subgraph folders
	Subgraphs  string `json:"subgraphs" envconfig:"SUBGRAPHS"`
	ManagerURL string `json:"manager_url" envconfig:"MANAGER_URL" default:"ws://0.0.0.0:8085/runner"`
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
