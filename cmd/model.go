package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshep3087/norav/pihole"
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
	applicationList list.Model

	metadata metadata

	healthcheckInterval time.Duration

	// showSonarrDetail is a flag to indicate if the sonarr detailed view should be shown
	showSonarrDetail bool
	// sonarrSeries is the list of series from the Sonarr instance
	sonarrSeriesList list.Model

	// windowSize is the size of the terminal window
	windowSize windowSize
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

	case checkApplicationsMsg:
		log.Println("received checkApplicationsMsg")

		for _, item := range m.applicationList.Items() {
			log.Printf("Checking %s", item.FilterValue())
			defer log.Printf("Checked %s", item.FilterValue())
			app, ok := item.(Application)
			if !ok {
				log.Printf("Item %s is not an application", item.FilterValue())
				continue
			}

			cmd = app.FetchStatus()
			cmds = append(cmds, cmd)
		}

		cmds = append(cmds, m.checkApplications(m.healthcheckInterval))

		log.Println("Checked applications")

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.applicationList.SetSize(msg.Width-h, msg.Height-v)

		m.windowSize = windowSize{Width: msg.Width, Height: msg.Height}

		return m, nil

	// case SeriesResourceMsg:
	// 	m.showSonarrDetail = true

	// 	items := make([]list.Item, len(msg.SeriesResources))
	// 	for i := range msg.SeriesResources {
	// 		items[i] = &msg.SeriesResources[i]
	// 	}

	// 	m.sonarrSeriesList = list.New(items, list.NewDefaultDelegate(), 0, 0)

	// 	h, v := docStyle.GetFrameSize()
	// 	m.sonarrSeriesList.SetSize(m.windowSize.Width-h, m.windowSize.Height-v)

	// 	return m, nil

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
				pihole := m.applicationList.SelectedItem().(*pihole.Model)

				pihole.SetActive(true)
				return m, pihole.FetchPiHoleStats
			}

			// if m.applicationList.SelectedItem().FilterValue() == "Sonarr" {
			// 	return m, m.fetchSonarrSeries
			// }

		case "esc":
			m.state = StateNormal
			return m, nil
		}

		// cmd = m.applicationList.SetItems(appsToItems(m.applications))
		// cmds = append(cmds, cmd)

		cmd = m.checkApplications(m.healthcheckInterval)
		cmds = append(cmds, cmd)
	}

	if m.state == StateNormal {
		log.Println("Updating application list")
		m.applicationList, cmd = m.applicationList.Update(msg)
		cmds = append(cmds, cmd)
	}

	for _, item := range m.applicationList.Items() {
		app, ok := item.(tea.Model)
		if !ok {
			log.Printf("Item %s is not an tea.Model", item.FilterValue())
			continue
		}
		_, cmd = app.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) deactivateAll() {
	for _, item := range m.applicationList.Items() {
		app, ok := item.(Application)
		if !ok {
			continue
		}
		app.SetActive(false)
	}
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
		pihole := m.applicationList.SelectedItem().(*pihole.Model)
		return pihole.View()
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
	// allGood := true
	// for _, app := range m.applications {
	// 	if app.httpResp.status != http.StatusOK {
	// 		allGood = false
	// 		break
	// 	}
	// }

	// Create status bar
	statusBar := statusBarStyle.Render("good for now")

	// Append status bar to the view
	b.WriteString("\n" + statusBar + "\n")

	return b.String()
}

func (m *model) applicationsView() string {
	return docStyle.Render(m.applicationList.View())
}

func (m *model) checkApplications(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return checkApplicationsMsg{}
	})
}

type checkApplicationsMsg struct{}

// type SeriesResourceMsg sonarr.Series

// // fetchSonarrSeries fetches the series from the Sonarr instance
// func (m *model) fetchSonarrSeries() tea.Msg {
// 	sonarrCfg := m.applicationList.SelectedItem().(application)

// 	sonarrClient := sonarr.NewClient(sonarrCfg.URL, sonarrCfg.AuthKey)
// 	series, err := sonarrClient.GetSeries()
// 	if err != nil {
// 		log.Printf("Error fetching series from Sonarr: %s", err)
// 		return nil
// 	}

// 	return SeriesResourceMsg(*series)
// }