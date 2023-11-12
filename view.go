package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

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
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Total Queries: ") + detailDataStyle.Render(m.piHoleSummary.DNSQueries) + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Queries Blocked: ") + detailDataStyle.Render(m.piHoleSummary.AdsBlocked) + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Percentage Blocked: ") + detailDataStyle.Render(m.piHoleSummary.AdsPercentage+"%%") + "\n"))
		piHoleDetailBuilder.WriteString(fmt.Sprintf(detailHeaderStyle.Render("Domains on Adlist: ") + detailDataStyle.Render(m.piHoleSummary.DomainsBlocked) + "\n"))
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

func (m *model) applicationsView() string {
	rows := buildTableRows(m.applications)
	m.appTable.SetRows(rows)
	return baseStyle.Render(m.appTable.View()) + "\n"
}
