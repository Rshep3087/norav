package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
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
)

type httpResp struct {
	status int
	err    error
}

type application struct {
	name        string
	url         string
	description string
	httpResp    httpResp
}

type model struct {
	applications []application
	cursor       int
}

func initialModel() model {
	return model{
		applications: []application{
			{
				name:        "Pi-hole",
				url:         "http://192.168.2.49/admin/",
				description: "a dns sinkhole",
			},
			{
				name:        "Home Assistant",
				url:         "http://192.168.2.49:8123/",
				description: "home automation",
			},
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
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

		case "ctrl+t":
			return m, checkServer(m.applications[m.cursor].url)
		}

	case statusMsg:
		m.applications[m.cursor].httpResp.status = int(msg)
	}

	return m, nil
}

func (m model) View() string {
	ui := strings.Builder{}

	title := titleStyle.Render("Ryan's Homelab")

	ui.WriteString(title)

	var items []string
	for i, app := range m.applications {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">"
		}

		s := fmt.Sprintf(
			"\n%s %s status: %d\n%s\n\n",
			cursor,
			app.name,
			app.httpResp.status,
			url(app.url),
		)
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

	return ui.String()
}

func checkServer(url string) tea.Cmd {
	return func() tea.Msg {
		c := &http.Client{Timeout: 10 * time.Second}
		res, err := c.Get(url)
		if err != nil {
			return errMsg{err}
		}

		return statusMsg(res.StatusCode)
	}
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type statusMsg int

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
