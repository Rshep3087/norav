package sonarr

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// messages for updates
type (
	// SonarrStatusMsg is a message that contains the status of the Sonarr instance
	SonarrStatusMsg string // "✅" or "❌"
)

type Model struct {
	// name of the Sonarr instance
	name string
	// host is the DNS or IP address of the Sonarr instance
	host string
	// apiKey is the API Token (available in the Sonarr web interface)
	apiKey string
	// description is the description of the Sonarr instance
	description string
	// active is a flag to indicate if the application is active
	active bool
	// healthStatus is the status of the Sonarr instance
	healthStatus string
	// list is the list of series from the Sonarr instance
	list list.Model
	// client is the sonarr client
	client *Client
}

func NewModel(cfg Config) *Model {
	m := Model{
		name:        cfg.Name,
		host:        cfg.Host,
		apiKey:      cfg.APIKey,
		description: cfg.Description,
		active:      false,
	}

	m.list = list.New(nil, list.NewDefaultDelegate(), 0, 0)

	client := NewClient(cfg.Host, cfg.APIKey)
	m.client = client

	return &m
}

func (m *Model) FilterValue() string { return m.name }
func (m *Model) Init() tea.Cmd       { return nil }
func (m *Model) Title() string       { return m.name }
func (m *Model) Description() string { return fmt.Sprintf("%s - %s", m.healthStatus, m.description) }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case SonarrStatusMsg:
		m.healthStatus = string(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return m.list.View()
}

func (m *Model) SetActive(b bool) { m.active = b }

func (m *Model) FetchStatus() tea.Cmd {
	return func() tea.Msg {
		log.Println("fetching sonarr status")

		var statusMsg SonarrStatusMsg

		resp, err := m.client.Health()
		if err != nil {
			statusMsg = "❌"
		} else {
			statusMsg = "✅"
		}

		log.Printf("Sonarr status: %+v", resp)

		return statusMsg
	}
}
