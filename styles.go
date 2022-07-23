package main

import "github.com/charmbracelet/lipgloss"

var (
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	listItem = lipgloss.NewStyle().
			MarginBottom(2).
			PaddingLeft(2).Render

	url = lipgloss.NewStyle().Foreground(special).Render

	titleStyle = lipgloss.NewStyle().
			MarginBottom(1).
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

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)
