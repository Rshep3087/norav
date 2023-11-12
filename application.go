package main

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/bubbles/table"
)

// application is a struct that holds the information for self-hosted application
// that is to be monitored
type application struct {
	Name              string `toml:"name"`
	URL               string `toml:"url"`
	Description       string `toml:"description"`
	httpResp          httpResp
	AuthHeader        string `toml:"authHeader"`
	AuthKey           string `toml:"authKey"`
	BasicAuthUsername string `toml:"basicAuthUsername"`
	BasicAuthPassword string `toml:"basicAuthPassword"`
}

func buildTableRows(apps []application) []table.Row {
	rows := make([]table.Row, len(apps))

	for i, app := range apps {
		statusEmoji := "❌"
		if app.httpResp.status == http.StatusOK {
			statusEmoji = "✅"
		}

		status := fmt.Sprintf("%s %d", statusEmoji, app.httpResp.status)

		row := table.Row{
			status,
			app.Name,
			app.URL,
		}

		rows[i] = row
	}

	return rows
}

func buildTableColumns() []table.Column {
	columns := []table.Column{
		{Title: "Status", Width: 10},
		{Title: "Name", Width: 20},
		{Title: "URL", Width: 70},
	}

	return columns
}

func buildApplicationTable(apps []application) table.Model {
	columns := buildTableColumns()
	rows := buildTableRows(apps)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
	)

	return t
}
