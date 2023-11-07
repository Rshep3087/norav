package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
)

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
	AdsBlocked       int     `json:"ads_blocked_today"`
	AdsPercentage    float64 `json:"ads_percentage_today"`
	ClientsEverSeen  int     `json:"clients_ever_seen"`
	DNSQueries       int     `json:"dns_queries_today"`
	DomainsBlocked   int     `json:"domains_being_blocked"`
	QueriesCached    int     `json:"queries_cached"`
	QueriesForwarded int     `json:"queries_forwarded"`
	Status           string  `json:"status"`
	UniqueClients    int     `json:"unique_clients"`
	UniqueDomains    int     `json:"unique_domains"`
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
// Answer types: 1 = blocked by gravity.list, 2 = forwarded to upstream server, 3 = answered by local cache, 4 = blocked by wildcard blocking
type PiHQueries struct {
	Data [][]string `json:"data"`
}

// Get performes API request. Returns slice of bytes.
func (ph *PiHConnector) Get(endpoint string) []byte {
	var requestString = "http://" + ph.Host + "/admin/api.php?" + endpoint
	if ph.Token != "" {
		requestString += "&auth=" + ph.Token
	}

	resp, err := http.Get(requestString)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
	fmt.Printf("- Blocked Domains: %d\n", ph.AdsBlocked)
	fmt.Printf("- Blocked Percentage: %.2f%%\n", ph.AdsPercentage)
	fmt.Printf("- Queries: %d\n", ph.DNSQueries)
	fmt.Printf("- Clients Ever Seen: %d\n", ph.ClientsEverSeen)
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
