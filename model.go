package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
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
	// showPiHoleDetail is a flag to indicate if the pi hole detailed view should be shown
	showPiHoleDetail bool
	// piHoleSummary store Pi-hole DNS statistics
	piHoleSummary PiHSummary
	// client is the http client used for making calls to the applications
	client *http.Client

	applicationList list.Model
	piHoleTable     table.Model
}

func (m model) Init() tea.Cmd {
	m.piHoleTable = table.New()
	m.piHoleTable.SetColumns([]table.Column{
		{Title: "Metric", Width: 20},
		{Title: "Value", Width: 20},
	})
	m.piHoleTable.SetWidth(100)
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
		log.Printf("key pressed: %s", msg.String())

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

		case "j", "down":
			if m.showPiHoleDetail {
				m.piHoleTable, cmd = m.piHoleTable.Update(msg)
				return m, cmd
			}
		case "k", "up":
			if m.showPiHoleDetail {
				m.piHoleTable, cmd = m.piHoleTable.Update(msg)
				return m, cmd
			}

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

		cmd = m.applicationList.SetItems(appsToItems(m.applications))

		return m, tea.Batch(cmd, m.checkApplications(m.healthcheckInterval))
	}

	m.applicationList, cmd = m.applicationList.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// Apply titleStyle to the title and add it to the top of the view
	title := titleStyle.Render(m.metadata.title)
	b.WriteString(title + "\n\n")

	log.Printf("showPiHoleDetail: %v", m.showPiHoleDetail)

	if m.showPiHoleDetail {
		// Update the table with Pi-hole statistics
		m.piHoleTable.SetRows([]table.Row{
			{"Status", m.piHoleSummary.Status},
			{"Total Queries", m.piHoleSummary.DNSQueries},
			{"Queries Blocked", m.piHoleSummary.AdsBlocked},
			{"Percentage Blocked", m.piHoleSummary.AdsPercentage + "%"},
			{"Domains on Adlist", m.piHoleSummary.DomainsBlocked},
			{"Unique Domains", m.piHoleSummary.UniqueDomains},
			{"Queries Cached", m.piHoleSummary.QueriesCached},
			{"Clients", m.piHoleSummary.ClientsEverSeen},
		})

		log.Printf("piHoleSummary: %+v", m.piHoleSummary)

		// Render the table
		return detailHeaderStyle.Render("Pi-hole Detailed View") + "\n\n" + m.piHoleTable.View()
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

func (m *model) checkApplications(d time.Duration) tea.Cmd {
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
func (m *model) fetchPiHoleStats() PiHSummary {
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
