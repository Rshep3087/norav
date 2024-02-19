package pihole

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	Host        string `toml:"host"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	APIKey      string `toml:"apiKey"`
}

type Model struct {
	name         string
	healthStatus string
	table        table.Model
	cfg          Config
}

// fetchPiHoleStats fetches statistics from the Pi-hole instance
func (m *Model) FetchPiHoleStats() tea.Msg {
	log.Println("Fetching new Pi-hole stats")
	defer log.Println("Fetched new Pi-hole stats")

	// Set up the Pi-hole connector with the URL and AuthKey from the config
	piHoleConnector := PiHConnector{
		Host:  m.cfg.Host,
		Token: m.cfg.APIKey,
	}
	stats := piHoleConnector.Summary()

	return piHoleSummaryMsg{summary: stats}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type piHoleSummaryMsg struct {
	summary PiHSummary
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case piHoleSummaryMsg:
		log.Println("Updating Pi-hole stats")
		defer log.Println("Updated Pi-hole stats")
		m.table.SetRows([]table.Row{
			{"Ads Blocked", msg.summary.AdsBlocked},
			{"Ads Percentage", msg.summary.AdsPercentage},
			{"Clients Ever Seen", msg.summary.ClientsEverSeen},
			{"DNS Queries", msg.summary.DNSQueries},
			{"Domains Blocked", msg.summary.DomainsBlocked},
		})
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	return m.table.View()
}

func NewApplication(cfg Config) Model {
	m := Model{
		cfg: cfg,
	}
	columns := []table.Column{
		{Title: "Metric", Width: 20},
		{Title: "Value", Width: 20},
	}

	m.table = table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
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
	m.table.SetStyles(s)

	return m
}

func (a *Model) HealthCheck() error {

	a.healthStatus = "Looking good..."

	return nil
}

func (a *Model) HealthStatus() string { return "" }

func (a *Model) Name() string { return a.name }

// PiHConnector represents base API connector type.
// Host: DNS or IP address of your Pi-Hole
// Token: API Token (see /etc/pihole/setupVars.conf)
type PiHConnector struct {
	Host  string
	Token string
}

// PiHType coitains Pi-Hole backend type (PHP or FTL).
type PiHType struct {
	Type string `json:"type"`
}

// PiHVersion contains Pi-Hole API version.
type PiHVersion struct {
	Version float32 `json:"version"`
}

// PiHSummary contains Pi-Hole summary data.
type PiHSummary struct {
	AdsBlocked       string `json:"ads_blocked_today"`
	AdsPercentage    string `json:"ads_percentage_today"`
	ClientsEverSeen  string `json:"clients_ever_seen"`
	DNSQueries       string `json:"dns_queries_today"`
	DomainsBlocked   string `json:"domains_being_blocked"`
	QueriesCached    string `json:"queries_cached"`
	QueriesForwarded string `json:"queries_forwarded"`
	Status           string `json:"status"`
	UniqueClients    string `json:"unique_clients"`
	UniqueDomains    string `json:"unique_domains"`
}

// PiHTimeData represents statistics over time.
// Each record contains number of queries/blocked ads within 10min timeframe.
type PiHTimeData struct {
	AdsOverTime     map[string]int `json:"ads_over_time"`
	DomainsOverTime map[string]int `json:"domains_over_time"`
}

// PiHTopItems contains top queries and top blocked domains.
// Format: "DNS": Frequency
type PiHTopItems struct {
	Queries map[string]int `json:"top_queries"`
	Blocked map[string]int `json:"top_ads"`
}

// PiHTopClients represents Pi-Hole client IPs with corresponding number of requests.
type PiHTopClients struct {
	Clients map[string]int `json:"top_sources"`
}

// PiHForwardDestinations represents number of queries that have been forwarded and the target.
type PiHForwardDestinations struct {
	Destinations map[string]float32 `json:"forward_destinations"`
}

// PiHQueryTypes contains DNS query type and number of queries.
type PiHQueryTypes struct {
	Types map[string]float32 `json:"querytypes"`
}

// PiHQueries contains all DNS queries.
// This is slice of slices of strings.
// Each slice contains: timestamp of query, type of query (IPv4, IPv6), requested DNS, requesting client, answer type.
type PiHQueries struct {
	Data [][]string `json:"data"`
}

// Get performes API request. Returns slice of bytes.
func (ph *PiHConnector) Get(endpoint string) []byte {
	log.Printf("Fetching data from Pi-hole API: %s", endpoint)
	defer log.Printf("Fetched data from Pi-hole API: %s", endpoint)

	var requestString = ph.Host + "/admin/api.php?" + endpoint
	if ph.Token != "" {
		requestString += "&auth=" + ph.Token
	}

	log.Printf("Requesting: %s", requestString)

	resp, err := http.Get(requestString)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body
}

// Type returns Pi-Hole API type as a PiHType object.
func (ph *PiHConnector) Type() PiHType {
	bs := ph.Get("type")
	s := &PiHType{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Version returns Pi-Hole API version as an object.
func (ph *PiHConnector) Version() PiHVersion {
	bs := ph.Get("version")
	s := &PiHVersion{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Summary returns statistics in formatted style.
func (ph *PiHConnector) Summary() PiHSummary {
	bs := ph.Get("summary")

	s := &PiHSummary{}
	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// TimeData returns PiHTimeData object which contains requests statistics.
func (ph *PiHConnector) TimeData() PiHTimeData {
	bs := ph.Get("overTimeData10mins")
	s := &PiHTimeData{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Top returns top blocked and requested domains.
func (ph *PiHConnector) Top(n int) PiHTopItems {
	bs := ph.Get("topItems=" + strconv.Itoa(n))
	s := &PiHTopItems{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Clients returns top clients.
func (ph *PiHConnector) Clients(n int) PiHTopClients {
	bs := ph.Get("topClients=" + strconv.Itoa(n))
	s := &PiHTopClients{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// ForwardDestinations returns forward destinations (DNS servers).
func (ph *PiHConnector) ForwardDestinations() PiHForwardDestinations {
	bs := ph.Get("getForwardDestinations")
	s := &PiHForwardDestinations{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// QueryTypes returns DNS query type and frequency as a PiHQueryTypes object.
func (ph *PiHConnector) QueryTypes() PiHQueryTypes {
	bs := ph.Get("getQueryTypes")
	s := &PiHQueryTypes{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Queries returns all DNS queries as a PiHQueries object.
func (ph *PiHConnector) Queries() PiHQueries {
	bs := ph.Get("getAllQueries")
	s := &PiHQueries{}

	err := json.Unmarshal(bs, s)
	if err != nil {
		log.Fatal(err)
	}
	return *s
}

// Enable enables Pi-Hole server.
func (ph *PiHConnector) Enable() error {
	bs := ph.Get("enable")
	resp := make(map[string]string)

	err := json.Unmarshal(bs, &resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp["status"] != "enabled" {
		return errors.New("something went wrong")
	}
	return nil
}

// Disable disables Pi-Hole server permanently.
func (ph *PiHConnector) Disable() error {
	bs := ph.Get("disable")
	resp := make(map[string]string)

	err := json.Unmarshal(bs, &resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp["status"] != "disabled" {
		return errors.New("something went wrong")
	}
	return nil
}

// RecentBlocked returns string with the last blocked DNS record.
func (ph *PiHConnector) RecentBlocked() string {
	bs := ph.Get("recentBlocked")
	return string(bs)
}

// Show returns 24h Summary of PiHole System.
func (ph *PiHSummary) Show() {
	fmt.Println("=== 24h Summary:")
	fmt.Printf("- Blocked Domains: %s\n", ph.AdsBlocked)
	fmt.Printf("- Blocked Percentage: %s\n", ph.AdsPercentage)
	fmt.Printf("- Queries: %s\n", ph.DNSQueries)
	fmt.Printf("- Clients Ever Seen: %s\n", ph.ClientsEverSeen)
}

// ShowBlocked returns sorted top Blocked domains over last 24h.
func (ph *PiHTopItems) ShowBlocked() {
	reverseMapBlocked := make(map[int]string)
	var freqBlocked []int

	for k, v := range ph.Blocked {
		reverseMapBlocked[v] = k
		freqBlocked = append(freqBlocked, v)
	}

	sort.Ints(freqBlocked)

	fmt.Println("=== Blocked domains over last 24h:")
	for i := len(freqBlocked) - 1; i >= 0; i-- {
		fmt.Printf("- %s : %d\n", reverseMapBlocked[freqBlocked[i]], freqBlocked[i])
	}
}

// ShowQueries returns sorted top queries over last 24h.
func (ph *PiHTopItems) ShowQueries() {
	reverseMapQueries := make(map[int]string)
	var freqQueries []int

	for k, v := range ph.Queries {
		reverseMapQueries[v] = k
		freqQueries = append(freqQueries, v)
	}

	sort.Ints(freqQueries)

	fmt.Println("=== Queries over last 24h:")
	for i := len(freqQueries) - 1; i >= 0; i-- {
		fmt.Printf("- %s : %d\n", reverseMapQueries[freqQueries[i]], freqQueries[i])
	}
}

// Show returns sorted top clients over last 24h.
func (ph *PiHTopClients) Show() {
	reverseMapClients := make(map[int]string)
	var freqClients []int

	for k, v := range ph.Clients {
		reverseMapClients[v] = k
		freqClients = append(freqClients, v)
	}

	sort.Ints(freqClients)

	fmt.Println("=== Clients over last 24h:")
	for i := len(freqClients) - 1; i >= 0; i-- {
		fmt.Printf("- %s : %d\n", reverseMapClients[freqClients[i]], freqClients[i])
	}
}
