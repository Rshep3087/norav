package pihole

// Config represents the configuration for the Pi-hole client.
type Config struct {
	// Host: DNS or IP address of your Pi-Hole
	Host string `toml:"host"`
	// Name: Name of the Pi-Hole instance
	Name string `toml:"name"`
	// Description: Description of the Pi-Hole instance
	Description string `toml:"description"`
	// APIKey is the API Token (available in the Pi-Hole web interface)
	APIKey string `toml:"apiKey"`
}
