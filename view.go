package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	// special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Padding(0, 1).
				Width(100)

	detailDataStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			PaddingTop(2).
			PaddingLeft(2).
			PaddingBottom(1).
			Width(22)

	titleStyle = lipgloss.NewStyle().
			MarginBottom(1).
			Align(lipgloss.Left).
			Background(lipgloss.Color("#FF5F87")).
			Width(100)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
)

func (m model) View() string {
	var b strings.Builder

	// Apply titleStyle to the title and add it to the top of the view
	title := titleStyle.Render(m.metadata.title)
	b.WriteString(title + "\n\n")

	if m.showPiHoleDetail {
		// Render the detailed page for Pi-hole with actual statistics
		var piHoleDetailBuilder strings.Builder
		piHoleDetailBuilder.WriteString(detailHeaderStyle.Render("Pi-hole Detailed View") + "\n\n")
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Total Queries: ") + detailDataStyle.Render(m.piHoleStats.DNSQueries) + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Queries Blocked: ") + detailDataStyle.Render(m.piHoleStats.AdsBlocked) + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Percentage Blocked: ") + detailDataStyle.Render(m.piHoleStats.AdsPercentage+"%%") + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Domains on Adlist: ") + detailDataStyle.Render(m.piHoleStats.DomainsBlocked) + "\n"))
		return piHoleDetailBuilder.String()
	}

	b.WriteString(m.applicationsView())

	// Check if all applications are good
	allGood := true
	for _, app := range m.applications {
		if app.httpResp.status != http.StatusOK {
			allGood = false
			break
		}
	}

	// Create status bar
	var statusBar string
	if allGood {
		statusBar = statusBarStyle.Render("All good..")
	} else {
		statusBar = statusBarStyle.Render(m.metadata.status)
	}

	// Append status bar to the view
	b.WriteString("\n" + statusBar + "\n")

	return b.String()
}

func (m model) applicationsView() string {

	columns := []table.Column{
		{Title: "Status", Width: 10},
		{Title: "Name", Width: 20},
		{Title: "URL", Width: 70},
	}

	rows := []table.Row{
		{"", "", ""},
	}

	for _, app := range m.applications {
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

		rows = append(rows, row)
	}

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
		Bold(false)
	t.SetStyles(s)

	return t.View() + "\n"
}
