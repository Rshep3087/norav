package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/rshep3087/norav/pihole"
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

func main() {
	fs := ff.NewFlagSet("norav")

	var (
		config = fs.String('c', "config", ".norav.toml", "path to config file")
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

	// ====================================================================
	// debug logging

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	appList := list.New(
		appsToItems(cfg.Applications),
		list.NewDefaultDelegate(),
		0,
		0,
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

	initialModel := model{
		applications: cfg.Applications,
		metadata: metadata{
			title:  cfg.Title,
			status: "loading...",
		},
		httpClient:          httpClient,
		healthcheckInterval: time.Duration(cfg.HealthCheckInterval) * time.Second,
		applicationList:     appList,
		pihole: pihole.NewApplication(
			pihole.Config{
				Host:   "http://hp-ubuntu",
				APIKey: "8bda4efbe21b7ea71d80fdf5eb8d4258cbc8ef1317e4eaa9f471ac6f4ca3b086",
			},
		),
	}

	initialModel.applicationList.Title = cfg.Title

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
