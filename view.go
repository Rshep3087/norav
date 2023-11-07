package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	listItem = lipgloss.NewStyle().
			MarginBottom(2).
			PaddingLeft(2).Render

	url = lipgloss.NewStyle().Foreground(special).Render

	titleStyle = lipgloss.NewStyle().
			MarginBottom(1).
			Align(lipgloss.Center).
			Width(100)

		// Status Bar.

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

func (m model) View() string {
	var b strings.Builder

	// Apply titleStyle to the title and add it to the top of the view
	title := titleStyle.Render(m.metadata.title)
	b.WriteString(title + "\n\n")

	b.WriteString(m.applicationsView())

	return b.String()
}

func (m model) applicationsView() string {
	var b strings.Builder

	for i, app := range m.applications {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">"
		}

		statusEmoji := "❌"
		if app.httpResp.status == http.StatusOK {
			statusEmoji = "✅"
		}

		// Apply listItem style to the entire line and url style to the URL
		line := fmt.Sprintf(
			"%s %s %s status: %d %s",
			cursor,
			statusEmoji,
			app.Name,
			app.httpResp.status,
			app.URL,
		)
		styledLine := listItem(line)
		styledURL := url(app.URL)
		// Replace the URL in the line with the styled URL
		styledLineWithStyledURL := strings.Replace(styledLine, app.URL, styledURL, 1)

		b.WriteString(styledLineWithStyledURL)
	}

	return b.String()
}
