package main

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/bubbles/list"
)

// application is a struct that holds the information for self-hosted application
// that is to be monitored
type application struct {
	Name              string `toml:"name"`
	URL               string `toml:"url"`
	Descript          string `toml:"description"`
	httpResp          httpResp
	AuthHeader        string `toml:"authHeader"`
	AuthKey           string `toml:"authKey"`
	BasicAuthUsername string `toml:"basicAuthUsername"`
	BasicAuthPassword string `toml:"basicAuthPassword"`
}

// Methods to satisfy the list.DefaultItem interface
func (a application) Title() string {
	statusEmoji := "❌"
	if a.httpResp.status == http.StatusOK {
		statusEmoji = "✅"
	}
	return fmt.Sprintf("%s %s", statusEmoji, a.Name)
}
func (a application) Description() string {
	return fmt.Sprintf("%s - %s", a.Descript, a.URL)
}
func (a application) FilterValue() string { return a.Name }

func appsToItems(apps []application) []list.Item {
	items := make([]list.Item, len(apps))
	for i, app := range apps {
		items[i] = app
	}
	return items
}
