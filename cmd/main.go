package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshep3087/norav/config"
	"github.com/rshep3087/norav/pihole"
	"github.com/rshep3087/norav/sonarr"

	"github.com/urfave/cli/v3"
)

type metadata struct {
	title  string
	status string
}

func Execute() error {
	cmd := &cli.Command{
		Name:  "norav",
		Usage: "norav is a terminal dashboard for monitoring the health of your homelab applications",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "path to config file",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cfg, err := config.LoadFile(c.String("config"))
			if err != nil {
				return err
			}

			if len(os.Getenv("DEBUG")) > 0 {
				f, err := tea.LogToFile("debug.log", "debug")
				if err != nil {
					fmt.Println("fatal:", err)
					os.Exit(1)
				}
				defer f.Close()
			}

			var listItems []list.Item
			if cfg.PiHole != nil {
				log.Println("adding pihole")
				piholeApp := pihole.NewModel(*cfg.PiHole)
				listItems = append(listItems, piholeApp)
			}

			if cfg.Sonarr != nil {
				log.Println("adding sonarr")
				sonarrApp := sonarr.NewModel(*cfg.Sonarr)
				listItems = append(listItems, sonarrApp)
			}

			appList := list.New(
				listItems,
				list.NewDefaultDelegate(),
				0,
				0,
			)

			initialModel := model{
				metadata: metadata{
					title:  cfg.Title,
					status: "loading...",
				},
				healthcheckInterval: time.Duration(cfg.HealthCheckInterval) * time.Second,
				applicationList:     appList,
			}

			initialModel.applicationList.Title = cfg.Title

			p := tea.NewProgram(initialModel, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		return err
	}

	return nil
}
