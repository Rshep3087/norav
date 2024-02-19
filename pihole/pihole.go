package pihole

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// msg types for Update
type (
	// piHoleStatusMsg is a message that contains the status of the Pi-hole instance
	piHoleStatusMsg string // "✅" or "❌"

	// piHoleSummaryMsg is a message that contains the summary of the Pi-hole instance
	piHoleSummaryMsg struct {
		summary PiHSummary
	}
)

// Model represents the Pi-hole model
type Model struct {
	// name is the name of the Pi-hole instance
	name string
	// host is the DNS or IP address of the Pi-hole instance
	host string
	// apiKey is the API Token (available in the Pi-Hole web interface)
	apiKey string
	// description is the description of the Pi-hole instance
	description string
	// active is a flag to indicate if the application is active
	active bool
	// healthStatus is the status of the Pi-hole instance
	healthStatus string
	// table is the table to display the Pi-hole summary
	table table.Model
}

// NewModel returns a new Pi-hole model
func NewModel(cfg Config) *Model {
	m := Model{
		name:        "Pi-hole",
		host:        cfg.Host,
		apiKey:      cfg.APIKey,
		description: "Network-wide ad blocking via your own Linux hardware",
		active:      false,
	}
	columns := []table.Column{
		{Title: "Metric", Width: 20},
		{Title: "Value", Width: 20},
	}

	m.table = table.New(
		table.WithColumns(columns),
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
	m.table.SetStyles(s)

	return &m
}

// Title returns the title of the Pi-hole instance
// Satisfaction of the tea.ItemDelegate interface
func (m *Model) Title() string       { return m.name }
func (m *Model) Description() string { return fmt.Sprintf("%s - %s", m.healthStatus, m.description) }

// fetchPiHoleStats fetches statistics from the Pi-hole instance
func (m *Model) FetchPiHoleStats() tea.Msg {
	// Set up the Pi-hole connector with the URL and AuthKey from the config
	piHoleConnector := PiHConnector{
		Host:  m.host,
		Token: m.apiKey,
	}
	stats, _ := piHoleConnector.Summary()

	return piHoleSummaryMsg{summary: stats}
}

// Init returns the initial command of the Pi-hole model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update updates the Pi-hole model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case piHoleSummaryMsg:
		log.Println("Updating Pi-hole stats")
		defer log.Println("Updated Pi-hole stats")
		m.table.SetRows([]table.Row{
			{"Ads Blocked", msg.summary.AdsBlocked},
			{"Ads Percentage", msg.summary.AdsPercentage},
			{"Clients Ever Seen", msg.summary.ClientsEverSeen},
			{"DNS Queries", msg.summary.DNSQueries},
			{"Domains Blocked", msg.summary.DomainsBlocked},
		})

		m.table.Focus()
		return m, nil

	case piHoleStatusMsg:
		log.Println("Updating Pi-hole status")
		defer log.Println("Updated Pi-hole status")

		log.Printf("Pi-hole status: %s", string(msg))
		m.healthStatus = string(msg)
	}

	if m.active {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View returns the view of the Pi-hole model
func (m *Model) View() string {
	return m.table.View()
}

// FilterValue returns the name of the Pi-hole instance
// Satisfaction of the tea.Item interface
func (m *Model) FilterValue() string { return m.name }

// FetchStatus fetches the status of the Pi-hole instance
// and returns a message with the status.
func (m *Model) FetchStatus() tea.Cmd {
	return func() tea.Msg {
		log.Println("Fetching Pi-hole status")
		piHoleConnector := PiHConnector{
			Host:  m.host,
			Token: m.apiKey,
		}
		var msg piHoleStatusMsg

		_, err := piHoleConnector.Type()
		if err != nil {
			msg = "❌"
		} else {
			msg = "✅"
		}

		return msg
	}
}

// SetActive sets the active flag
func (a *Model) SetActive(b bool) { a.active = b }
