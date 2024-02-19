package config

import "github.com/rshep3087/norav/pihole"

type Config struct {
	PiHole pihole.Config `toml:"pihole"`
	// Sonarr Sonarr `toml:"sonarr"`
}
