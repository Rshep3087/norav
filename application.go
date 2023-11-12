package main

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
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
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return t
}
