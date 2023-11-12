package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// config is a struct that holds homie configuration
type config struct {
	// Title of your homie dashboard
	Title string `toml:"title"`
	// Applications is a list of applications to be monitored
	Applications []application `toml:"applications"`
	// HealthCheckInterval is the interval in seconds to check the health of the applications
	HealthCheckInterval int `toml:"interval"`
}

// loadConfigFile loads the configuration file if it exists
func loadConfigFile(f string) (config, error) {
	if _, err := os.Stat(f); err != nil {
		return config{}, fmt.Errorf("config file %s does not exist: %w", f, err)
	}

	var cfg config
	_, err := toml.DecodeFile(f, &cfg)
	if err != nil {
		return config{}, fmt.Errorf("failed to decode config file %s: %w", f, err)
	}

	return cfg, nil
}
