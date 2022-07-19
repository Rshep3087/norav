package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	width = 96

	columnWidth = 30
)

var (
	subtle  = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	appStyle = lipgloss.NewStyle().
			MarginRight(2).
			Height(4).
			Width(columnWidth + 1)

	listItem = lipgloss.NewStyle().PaddingLeft(2).Render

	url = lipgloss.NewStyle().Foreground(special).Render

	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			Align(lipgloss.Center).
			Width(100)

		// Status Bar.

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))
)

type httpResp struct {
	status int
}

type config struct {
	Title        string
	Applications []application
}

type application struct {
	Name        string
	URL         string
	Description string
	httpResp    httpResp
}

type metadata struct {
	title  string
	status string
}

type model struct {
	applications []application
	cursor       int
	metadata     metadata

	client *http.Client
}

func (m model) GetAppURLs() []string {
	var urls []string
	for _, v := range m.applications {
		urls = append(urls, v.URL)
	}

	return urls
}

func (m model) Init() tea.Cmd {
	return m.checkServers()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.applications)-1 {
				m.cursor++
			}

		case "ctrl+h":
			return m, m.checkServers()
		}

	case statusMsg:
		m.metadata.status = "Looking good..."
		for i, app := range m.applications {
			m.applications[i].httpResp.status = msg[app.URL]
			if m.applications[i].httpResp.status != http.StatusOK {
				m.metadata.status = fmt.Sprintf("%s might be having issues...", app.Name)
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	ui := strings.Builder{}

	title := titleStyle.Render(m.metadata.title)

	ui.WriteString(title)

	var items []string
	for i, app := range m.applications {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">"
		}

		s := fmt.Sprintf(
			"%s %s status: %d\n%s\n\n",
			cursor,
			app.Name,
			app.httpResp.status,
			url(app.URL),
		)

		if app.httpResp.status == 0 {
			s = fmt.Sprintf(
				"%s %s status: \n%s\n\n",
				cursor,
				app.Name,
				url(app.URL),
			)

		}

		items = append(items, listItem(s))
	}

	apps := lipgloss.JoinHorizontal(lipgloss.Top,
		appStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				items...,
			),
		),
	)

	ui.WriteString(apps)

	{
		w := lipgloss.Width

		statusKey := statusStyle.Render("STATUS")
		encoding := encodingStyle.Render("UTF-8")
		fishCake := fishCakeStyle.Render("🍥 Fish Cake")
		statusVal := statusText.Copy().
			Width(width - w(statusKey) - w(encoding) - w(fishCake)).
			Render(m.metadata.status)

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			encoding,
			fishCake,
		)

		ui.WriteString(statusBarStyle.Width(width).Render(bar))
	}

	return ui.String()
}

func loadConfigFile() (config, error) {
	f := ".homie.toml"
	if _, err := os.Stat(f); err != nil {
		return config{}, err
	}

	var cfg config

	_, err := toml.DecodeFile(f, &cfg)
	if err != nil {
		return config{}, err
	}
	return cfg, nil
}

func (m model) checkServers() tea.Cmd {
	return func() tea.Msg {

		msg := make(statusMsg)

		for _, app := range m.applications {
			res, _ := m.client.Get(app.URL)
			msg[app.URL] = res.StatusCode
		}

		return msg
	}
}

type statusMsg map[string]int

func main() {
	// ====================================================================
	// load config file
	cfg, err := loadConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	// ====================================================================
	// clients
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	initialModel := model{
		applications: cfg.Applications,
		metadata: metadata{
			title:  cfg.Title,
			status: "loading...",
		},
		client: httpClient,
	}

	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
