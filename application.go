package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Application interface {
	list.Item

	// FetchStatus fetches the status of the application
	FetchStatus() tea.Cmd
}

// // Methods to satisfy the list.DefaultItem interface
// func (a application) Title() string {
// 	statusEmoji := "❌"
// 	if a.httpResp.status == http.StatusOK {
// 		statusEmoji = "✅"
// 	}
// 	return fmt.Sprintf("%s %s", statusEmoji, a.Name)
// }
// func (a application) Description() string {
// 	return fmt.Sprintf("%s - %s", a.Descript, a.URL)
// }
// func (a application) FilterValue() string { return a.Name }

// func appsToItems(apps []application) []list.Item {
// 	items := make([]list.Item, len(apps))
// 	for i, app := range apps {
// 		items[i] = app
// 	}
// 	return items
// }
