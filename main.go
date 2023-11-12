package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

const (
	width = 96
)

type statusMsg map[string]int

type httpResp struct {
	status int
}

type metadata struct {
	title  string
	status string
}

// model is the bubbletea model
type model struct {
	// applications is a list of applications to be monitored
	applications        []application
	cursor              int
	metadata            metadata
	healthcheckInterval time.Duration
	viewport            viewport.Model
	// showPiHoleDetail is a flag to indicate if the pi hole detailed view should be shown
	showPiHoleDetail bool
	// piHoleSummary store Pi-hole DNS statistics
	piHoleSummary PiHSummary
	appTable      table.Model
	client        *http.Client
}

func (m model) Init() tea.Cmd {
	m.viewport = viewport.Model{Width: width, Height: 10}
	m.viewport.YPosition = 0
	return m.checkServers(10 * time.Millisecond)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewport.YOffset {
					m.viewport.YOffset--
				}
			}
		case "down", "j":
			if m.cursor < len(m.applications)-1 {
				m.cursor++
				if m.cursor >= m.viewport.YOffset+m.viewport.Height {
					m.viewport.YOffset++
				}
			}
		case "enter":
			if m.applications[m.cursor].Name == "Pi-hole" {
				m.showPiHoleDetail = true
				m.piHoleSummary = m.fetchPiHoleStats()
			}
		case "esc":
			if m.showPiHoleDetail {
				m.showPiHoleDetail = false
			}
		default:
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
		return m, m.checkServers(m.healthcheckInterval)
	}

	m.appTable, cmd = m.appTable.Update(msg)

	return m, cmd
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

func (m model) checkServers(d time.Duration) tea.Cmd {
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

func main() {
	fs := ff.NewFlagSet("homie")

	var (
		config = fs.String('c', "config", ".homie.toml", "path to config file")
	)

	err := fs.Parse(os.Args[1:])
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
		os.Exit(0)
	case err != nil:
		log.Fatal(err)
	}

	// ====================================================================
	// load config file
	cfg, err := loadConfigFile(*config)
	if err != nil {
		log.Fatal(err)
	}

	// ====================================================================
	// clients
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	appTable := buildApplicationTable(cfg.Applications)

	initialModel := model{
		applications: cfg.Applications,
		metadata: metadata{
			title:  cfg.Title,
			status: "loading...",
		},
		client:              httpClient,
		healthcheckInterval: time.Duration(cfg.HealthCheckInterval) * time.Second,
		appTable:            appTable,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
