package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshep3087/norav/pihole"
	"github.com/rshep3087/norav/sonarr"
)

type noravState int

const (
	// StateNormal is the default state
	StateNormal noravState = iota
	// StatePiHole is the state for the Pi-hole detailed view
	StatePiHole
	// StateSonarr is the state for the Sonarr detailed view
	StateSonarr
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Padding(0, 1).
				Width(100)

	detailDataStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			PaddingTop(2).
			PaddingLeft(2).
			PaddingBottom(1).
			Width(22)

	titleStyle = lipgloss.NewStyle().
			MarginBottom(1).
			Align(lipgloss.Left).
			Background(lipgloss.Color("#FF5F87")).
			Width(100)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
)

// model is the bubbletea model
type model struct {
	// state is the current state of the application
	state noravState
	// applications is a list of applications to be monitored
	applications    []application
	applicationList list.Model

	metadata metadata

	healthcheckInterval time.Duration
	// client is the http client used for making calls to the applications
	httpClient *http.Client

	// showSonarrDetail is a flag to indicate if the sonarr detailed view should be shown
	showSonarrDetail bool
	// sonarrSeries is the list of series from the Sonarr instance
	sonarrSeriesList list.Model

	// windowSize is the size of the terminal window
	windowSize windowSize

	// pihole is the Pi-hole model
	pihole pihole.Model
}

type windowSize struct {
	Width  int
	Height int
}

func (m model) Init() tea.Cmd {
	return m.checkApplications(10 * time.Millisecond)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.applicationList.SetSize(msg.Width-h, msg.Height-v)

		m.windowSize = windowSize{Width: msg.Width, Height: msg.Height}

		return m, nil

	case SeriesResourceMsg:
		m.showSonarrDetail = true

		items := make([]list.Item, len(msg.SeriesResources))
		for i := range msg.SeriesResources {
			items[i] = &msg.SeriesResources[i]
		}

		m.sonarrSeriesList = list.New(items, list.NewDefaultDelegate(), 0, 0)

		h, v := docStyle.GetFrameSize()
		m.sonarrSeriesList.SetSize(m.windowSize.Width-h, m.windowSize.Height-v)

		return m, nil

	case tea.KeyMsg:
		log.Printf("key pressed: %s", msg.String())

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			m.deactivateAll()
			// change the state to the selected application
			m.state = chooseState(m.applicationList.SelectedItem())

			log.Printf("Selected application: %s", m.applicationList.SelectedItem().FilterValue())
			if m.state == StatePiHole {
				m.pihole.SetActive(true)
				return m, m.pihole.FetchPiHoleStats
			}

			if m.applicationList.SelectedItem().FilterValue() == "Sonarr" {
				return m, m.fetchSonarrSeries
			}

		case "esc":
			m.state = StateNormal
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

		cmd = m.applicationList.SetItems(appsToItems(m.applications))
		cmds = append(cmds, cmd)

		cmd = m.checkApplications(m.healthcheckInterval)
		cmds = append(cmds, cmd)

	}

	m.pihole, cmd = m.pihole.Update(msg)
	cmds = append(cmds, cmd)

	if m.state == StateNormal {
		log.Println("Updating application list")
		m.applicationList, cmd = m.applicationList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) deactivateAll() {
	m.pihole.SetActive(false)
}

func chooseState(item list.Item) noravState {
	switch item.FilterValue() {
	case "Pi-hole":
		return StatePiHole
	case "Sonarr":
		return StateSonarr
	default:
		return StateNormal
	}
}

func (m model) View() string {

	switch m.state {
	case StatePiHole:
		return m.pihole.View()
	}

	// if m.showPiHoleDetail {
	// 	// Update the table with Pi-hole statistics
	// 	m.piHoleTable.SetRows([]table.Row{
	// 		{"Status", m.piHoleSummaryCache.Summary.Status},
	// 		{"Total Queries", m.piHoleSummaryCache.Summary.DNSQueries},
	// 		{"Queries Blocked", m.piHoleSummaryCache.Summary.AdsBlocked},
	// 		{"Percentage Blocked", m.piHoleSummaryCache.Summary.AdsPercentage + "%"},
	// 		{"Domains on Adlist", m.piHoleSummaryCache.Summary.DomainsBlocked},
	// 		{"Unique Domains", m.piHoleSummaryCache.Summary.UniqueDomains},
	// 		{"Queries Cached", m.piHoleSummaryCache.Summary.QueriesCached},
	// 		{"Clients", m.piHoleSummaryCache.Summary.ClientsEverSeen},
	// 	})

	// 	// Render the table
	// 	return detailHeaderStyle.Render("Pi-hole Detailed View") + "\n\n" + m.piHoleTable.View()
	// }

	if m.showSonarrDetail {
		return docStyle.Render(m.sonarrSeriesList.View())
	}
	var b strings.Builder

	// Apply titleStyle to the title and add it to the top of the view
	title := titleStyle.Render(m.metadata.title)
	b.WriteString(title + "\n\n")

	b.WriteString(m.applicationsView())

	// Check if all applications are good
	allGood := true
	for _, app := range m.applications {
		if app.httpResp.status != http.StatusOK {
			allGood = false
			break
		}
	}

	// Create status bar
	var statusBar string
	if allGood {
		statusBar = statusBarStyle.Render("All good..")
	} else {
		statusBar = statusBarStyle.Render(m.metadata.status)
	}

	// Append status bar to the view
	b.WriteString("\n" + statusBar + "\n")

	return b.String()
}

func (m *model) applicationsView() string {
	return docStyle.Render(m.applicationList.View())
}

func (m *model) checkApplications(d time.Duration) tea.Cmd {
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

			res, err := m.httpClient.Do(req)
			if err != nil {
				log.Printf("Error fetching %s: %s", app.URL, err)
				msg[app.URL] = 0
				continue
			}
			if res.StatusCode != http.StatusOK {
				log.Printf("Error fetching %s: %s", app.URL, res.Status)
			}
			msg[app.URL] = res.StatusCode
		}
		return msg
	})
}

type SeriesResourceMsg sonarr.Series

// fetchSonarrSeries fetches the series from the Sonarr instance
func (m *model) fetchSonarrSeries() tea.Msg {
	sonarrCfg := m.applicationList.SelectedItem().(application)

	sonarrClient := sonarr.NewClient(sonarrCfg.URL, sonarrCfg.AuthKey)
	series, err := sonarrClient.GetSeries()
	if err != nil {
		log.Printf("Error fetching series from Sonarr: %s", err)
		return nil
	}

	return SeriesResourceMsg(*series)
}
