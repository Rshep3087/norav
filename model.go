package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	docStyle = lipgloss.NewStyle().Margin(1, 2)

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

// model is the bubbletea model
type model struct {
	// applications is a list of applications to be monitored
	applications        []application
	metadata            metadata
	healthcheckInterval time.Duration
	viewport            viewport.Model
	// showPiHoleDetail is a flag to indicate if the pi hole detailed view should be shown
	showPiHoleDetail bool
	// piHoleSummary store Pi-hole DNS statistics
	piHoleSummary PiHSummary
	// client is the http client used for making calls to the applications
	client *http.Client

	applicationList list.Model
}

func (m model) Init() tea.Cmd {
	m.viewport = viewport.Model{Width: width, Height: 10}
	m.viewport.YPosition = 0
	return m.checkApplications(10 * time.Millisecond)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.applicationList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			log.Println("enter pressed")
			if m.applicationList.SelectedItem().FilterValue() == "Pi-hole" {
				m.showPiHoleDetail = true
				m.piHoleSummary = m.fetchPiHoleStats()
			}

		case "esc":
			if m.showPiHoleDetail {
				m.showPiHoleDetail = false
			}

			return m, nil
		}

	case statusMsg:
		m.metadata.status = "Looking good..."
		for i, app := range m.applications {
			m.applications[i].httpResp.status = msg[app.URL]
			switch m.applications[i].httpResp.status {
			case http.StatusOK:
				// do nothing
			default:
				m.metadata.status = fmt.Sprintf("%s might be having issues...", app.Name)
			}
		}

		m.applicationList.SetItems(appsToItems(m.applications))

		return m, m.checkApplications(m.healthcheckInterval)
	}

	m.applicationList, cmd = m.applicationList.Update(msg)
	return m, cmd
}

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
	return docStyle.Render(m.applicationList.View())
}

func (m model) checkApplications(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		msg := make(statusMsg)

		for _, app := range m.applications {
			req, err := http.NewRequest("GET", app.URL, nil)
			if err != nil {
				msg[app.URL] = 0
				continue
			}

			if app.AuthHeader != "" && app.AuthKey != "" {
				req.Header.Add(app.AuthHeader, app.AuthKey)
			}

			res, err := m.client.Do(req)
			if err != nil {
				msg[app.URL] = 0
				continue
			}
			msg[app.URL] = res.StatusCode
		}
		return msg
	})
}

// fetchPiHoleStats fetches statistics from the Pi-hole instance
func (m model) fetchPiHoleStats() PiHSummary {
	// Find the Pi-hole application configuration
	var piHoleApp application
	for _, app := range m.applications {
		if app.Name == "Pi-hole" {
			piHoleApp = app
			break
		}
	}

	piholeURL := strings.TrimPrefix(piHoleApp.URL, "http://") // Remove "http://" prefix if present
	piholeURL = strings.TrimSuffix(piholeURL, "/admin/")      // Remove "/admin" suffix if present

	// Set up the Pi-hole connector with the URL and AuthKey from the config
	piHoleConnector := PiHConnector{
		Host:  piholeURL,         // Remove "http://" prefix if present
		Token: piHoleApp.AuthKey, // Use the AuthKey as the token
	}
	stats := piHoleConnector.Summary()
	return stats
}
