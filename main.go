package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	selected     map[int]struct{}
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
		selected: make(map[int]struct{}),
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

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
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
	s := "All Applications\n\n"

	for i, app := range m.applications {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf(
			"%s [%s] %s status: %d\n%s\n\n",
			cursor,
			checked,
			app.name,
			app.httpResp.status,
			app.url,
		)
	}

	s += "\n Press q to quit.\n"

	return s
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
