package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rshep3087/norav/pihole"
	"github.com/rshep3087/norav/sonarr"
)

// config is a struct that holds norav configuration
type config struct {
	// Title of your norav dashboard
	Title string `toml:"title"`
	// HealthCheckInterval is the interval in seconds to check the health of the applications
	HealthCheckInterval int `toml:"interval"`

	// PiHole is the configuration for the Pi-hole client
	PiHole *pihole.Config `toml:"pihole"`
	// Sonarr is the configuration for the Sonarr client
	Sonarr *sonarr.Config `toml:"sonarr"`
}

// loadConfigFile loads the configuration file if it exists
func LoadFile(f string) (config, error) {
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
