package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

const (
	width = 96
)

type httpResp struct {
	status int
}

type config struct {
	Title               string        `toml:"title"`
	Applications        []application `toml:"applications"`
	HealthCheckInterval int           `toml:"interval"`
}

type application struct {
	Name              string `toml:"name"`
	URL               string `toml:"url"`
	Description       string `toml:"description"`
	httpResp          httpResp
	AuthHeader        string `toml:"authHeader"`
	AuthKey           string `toml:"authKey"`
	BasicAuthUsername string `toml:"basicAuthUsername"`
	BasicAuthPassword string `toml:"basicAuthPassword"`
}

type metadata struct {
	title  string
	status string
}

type model struct {
	applications        []application
	cursor              int
	metadata            metadata
	healthcheckInterval time.Duration
	viewport            viewport.Model

	client *http.Client
}

func (m model) Init() tea.Cmd {
	m.viewport = viewport.Model{Width: width, Height: 10}
	m.viewport.YPosition = 0
	return m.checkServers(10 * time.Millisecond)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	return m, nil
}

func loadConfigFile(f string) (config, error) {
	if _, err := os.Stat(f); err != nil {
		return config{}, fmt.Errorf("config file %s does not exist: %w", f, err)
	}

	var cfg config

	_, err := toml.DecodeFile(f, &cfg)
	if err != nil {
		return config{}, fmt.Errorf("failed to decode config file %s: %w", f, err)
	}
	return cfg, nil
}

type statusMsg map[string]int

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

	initialModel := model{
		applications: cfg.Applications,
		metadata: metadata{
			title:  cfg.Title,
			status: "loading...",
		},
		client:              httpClient,
		healthcheckInterval: time.Duration(cfg.HealthCheckInterval) * time.Second,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
