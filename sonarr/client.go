package sonarr

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

// Client represents the Sonarr client.
type Client struct {
	URL    string
	APIKey string
}

// NewClient creates a new Sonarr client.
func NewClient(url, apiKey string) *Client {
	return &Client{
		URL:    url,
		APIKey: apiKey,
	}
}

// GetSeries fetches the series from the Sonarr instance.
func (c *Client) GetSeries() (*Series, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", c.URL+"/api/v3/series?includeSeasonImages=false", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)

	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Request: %s", reqBytes)

	log.Println("GET", c.URL+"/api/v3/series?includeSeasonImages=false")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Response: %+v", resp.Status)

	decoded, err := decode[[]SeriesResource](resp)
	if err != nil {
		return nil, err
	}

	return &Series{SeriesResources: decoded}, nil
}

type HealthResponse struct {
	statusCode int
	status     string
}

func (c *Client) Health() (HealthResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", c.URL+"/api/v3/health", nil)
	if err != nil {
		return HealthResponse{}, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)

	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		return HealthResponse{}, err
	}

	log.Printf("Request: %s", reqBytes)

	log.Println("GET", c.URL+"/api/v3/health")
	resp, err := client.Do(req)
	if err != nil {
		return HealthResponse{}, err
	}

	return HealthResponse{statusCode: resp.StatusCode, status: resp.Status}, nil
}

func decode[T any](resp *http.Response) (T, error) {
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
