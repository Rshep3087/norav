package pihole

import (
	"log"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	Host        string `toml:"host"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	APIKey      string `toml:"apiKey"`
}

type Model struct {
	// active is a flag to indicate if the application is active
	active       bool
	name         string
	healthStatus string
	table        table.Model
	cfg          Config
}

// fetchPiHoleStats fetches statistics from the Pi-hole instance
func (m *Model) FetchPiHoleStats() tea.Msg {
	// Set up the Pi-hole connector with the URL and AuthKey from the config
	piHoleConnector := PiHConnector{
		Host:  m.cfg.Host,
		Token: m.cfg.APIKey,
	}
	stats := piHoleConnector.Summary()

	return piHoleSummaryMsg{summary: stats}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type piHoleSummaryMsg struct {
	summary PiHSummary
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
	}

	if m.active {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.table.View()
}

func NewApplication(cfg Config) Model {
	m := Model{
		cfg:    cfg,
		active: false,
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

	return m
}

func (a *Model) HealthCheck() error {

	a.healthStatus = "Looking good..."

	return nil
}

func (a *Model) HealthStatus() string { return "" }
func (a *Model) Name() string         { return a.name }

// SetActive sets the active flag
func (a *Model) SetActive(b bool) { a.active = b }
