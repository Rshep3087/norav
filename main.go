package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type application struct {
	name        string
	url         string
	description string
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
		}
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

		s += fmt.Sprintf("%s [%s] %s\n%s\n\n", cursor, checked, app.name, app.url)
	}

	s += "\n Press q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
