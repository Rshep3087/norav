package sonarr

// Config represents the configuration for the Sonarr client.
type Config struct {
	// Host: DNS or IP address of your Sonarr instance
	Host string `toml:"host"`
	// Name: Name of the Sonarr instance
	Name string `toml:"name"`
	// Description: Description of the Sonarr instance
	Description string `toml:"description"`
	// APIKey is the API Token (available in the Sonarr web interface)
	APIKey string `toml:"apiKey"`
}
