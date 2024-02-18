package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshep3087/norav/sonarr"
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
	// applications is a list of applications to be monitored
	applications    []application
	applicationList list.Model

	metadata metadata

	healthcheckInterval time.Duration
	// client is the http client used for making calls to the applications
	httpClient *http.Client

	// pi hole fields
	piHoleTable table.Model
	// piHoleSummaryCache stores the cached Pi-hole DNS statistics
	piHoleSummaryCache PiHSummaryCache
	// showPiHoleDetail is a flag to indicate if the pi hole detailed view should be shown
	showPiHoleDetail bool

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

// PiHSummaryCache is used to cache the PiHSummary for a duration
type PiHSummaryCache struct {
	Summary   PiHSummary
	Timestamp time.Time
}

func (m model) Init() tea.Cmd {
	m.piHoleTable = table.New()
	m.piHoleTable.SetColumns([]table.Column{
		{Title: "Metric", Width: 20},
		{Title: "Value", Width: 20},
	})
	m.piHoleTable.SetWidth(100)
	return m.checkApplications(10 * time.Millisecond)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.applicationList.SetSize(msg.Width-h, msg.Height-v)

		m.windowSize = windowSize{Width: msg.Width, Height: msg.Height}

		return m, nil

	case piHoleSummaryMsg:
		m.showPiHoleDetail = true
		m.piHoleSummaryCache.Summary = msg.summary
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
			log.Printf("Selected application: %s", m.applicationList.SelectedItem().FilterValue())
			if m.applicationList.SelectedItem().FilterValue() == "Pi-hole" {
				return m, m.fetchPiHoleStats
			}

			if m.applicationList.SelectedItem().FilterValue() == "Sonarr" {
				log.Println("Fetching Sonarr series")
				return m, m.fetchSonarrSeries
			}

		case "esc":
			if m.showPiHoleDetail {
				m.showPiHoleDetail = false
			}

			if m.showSonarrDetail {
				m.showSonarrDetail = false
			}

			return m, nil

		case "j", "down":
			if m.showPiHoleDetail {
				m.piHoleTable, cmd = m.piHoleTable.Update(msg)
				return m, cmd
			}
			if m.showSonarrDetail {
				m.sonarrSeriesList, cmd = m.sonarrSeriesList.Update(msg)
				return m, cmd
			}

		case "k", "up":
			if m.showPiHoleDetail {
				m.piHoleTable, cmd = m.piHoleTable.Update(msg)
				return m, cmd
			}

			if m.showSonarrDetail {
				m.sonarrSeriesList, cmd = m.sonarrSeriesList.Update(msg)

				return m, cmd
			}
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

		return m, tea.Batch(cmd, m.checkApplications(m.healthcheckInterval))
	}

	m.applicationList, cmd = m.applicationList.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// Apply titleStyle to the title and add it to the top of the view
	title := titleStyle.Render(m.metadata.title)
	b.WriteString(title + "\n\n")

	if m.showPiHoleDetail {
		// Update the table with Pi-hole statistics
		m.piHoleTable.SetRows([]table.Row{
			{"Status", m.piHoleSummaryCache.Summary.Status},
			{"Total Queries", m.piHoleSummaryCache.Summary.DNSQueries},
			{"Queries Blocked", m.piHoleSummaryCache.Summary.AdsBlocked},
			{"Percentage Blocked", m.piHoleSummaryCache.Summary.AdsPercentage + "%"},
			{"Domains on Adlist", m.piHoleSummaryCache.Summary.DomainsBlocked},
			{"Unique Domains", m.piHoleSummaryCache.Summary.UniqueDomains},
			{"Queries Cached", m.piHoleSummaryCache.Summary.QueriesCached},
			{"Clients", m.piHoleSummaryCache.Summary.ClientsEverSeen},
		})

		// Render the table
		return detailHeaderStyle.Render("Pi-hole Detailed View") + "\n\n" + m.piHoleTable.View()
	}

	if m.showSonarrDetail {
		log.Println("rendering Sonarr detailed view")
		return docStyle.Render(m.sonarrSeriesList.View())
	}

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

type piHoleSummaryMsg struct {
	summary PiHSummary
}

// fetchPiHoleStats fetches statistics from the Pi-hole instance
func (m *model) fetchPiHoleStats() tea.Msg {
	// Check if the cache is still valid
	if time.Since(m.piHoleSummaryCache.Timestamp) < 1*time.Minute {
		log.Println("Using cached Pi-hole stats")
		return piHoleSummaryMsg{summary: m.piHoleSummaryCache.Summary}
	}

	log.Println("Fetching new Pi-hole stats")

	// Cache is invalid or empty, fetch new data
	var piHoleApp application
	for _, app := range m.applications {
		if app.Name == "Pi-hole" {
			piHoleApp = app
			break
		}
	}

	piholeURL := strings.TrimPrefix(piHoleApp.URL, "http://") // Remove "http://" prefix if present
	piholeURL = strings.TrimSuffix(piholeURL, "/admin/")      // Remove "/admin" suffix if present

	// Set up the Pi-hole connector with the URL and AuthKey from the config
	piHoleConnector := PiHConnector{
		Host:  piholeURL,
		Token: piHoleApp.AuthKey,
	}
	stats := piHoleConnector.Summary()

	// Update the cache with the new data and timestamp
	m.piHoleSummaryCache = PiHSummaryCache{
		Summary:   stats,
		Timestamp: time.Now(),
	}

	return piHoleSummaryMsg{summary: stats}
}

type SeriesResourceMsg sonarr.Series

// fetchSonarrSeries fetches the series from the Sonarr instance
func (m *model) fetchSonarrSeries() tea.Msg {
	sonarrCfg := m.applicationList.SelectedItem().(application)

	sonarrClient := sonarr.NewClient(sonarrCfg.URL, sonarrCfg.AuthKey)
	series, err := sonarrClient.GetSeries()
	if err != nil {
		log.Println(err)
		return nil
	}

	return SeriesResourceMsg(*series)
}
