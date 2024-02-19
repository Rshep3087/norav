package pihole

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

func (m *Model) Init() tea.Cmd {
	return nil
}

type piHoleSummaryMsg struct {
	summary PiHSummary
}

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

func (m *Model) View() string {
	return m.table.View()
}

type piHoleStatusMsg string // "✅" or "❌"

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

type Config struct {
	Host        string `toml:"host"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	APIKey      string `toml:"apiKey"`
}
type Model struct {
	name        string
	host        string
	apiKey      string
	description string
	// active is a flag to indicate if the application is active
	active       bool
	healthStatus string
	table        table.Model
}

func NewApplication(cfg Config) *Model {
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
