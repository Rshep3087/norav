package sonarr

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type Series struct {
	SeriesResources []SeriesResource
}

type SeriesResource struct {
	ID                int               `json:"id"`
	MainTitle         string            `json:"title"`
	AlternateTitles   []AlternateTitles `json:"alternateTitles"`
	SortTitle         string            `json:"sortTitle"`
	Status            string            `json:"status"`
	Ended             bool              `json:"ended"`
	ProfileName       string            `json:"profileName"`
	Overview          string            `json:"overview"`
	NextAiring        time.Time         `json:"nextAiring"`
	PreviousAiring    time.Time         `json:"previousAiring"`
	Network           string            `json:"network"`
	AirTime           string            `json:"airTime"`
	Images            []Images          `json:"images"`
	OriginalLanguage  OriginalLanguage  `json:"originalLanguage"`
	RemotePoster      string            `json:"remotePoster"`
	Seasons           []Seasons         `json:"seasons"`
	Year              int               `json:"year"`
	Path              string            `json:"path"`
	QualityProfileID  int               `json:"qualityProfileId"`
	SeasonFolder      bool              `json:"seasonFolder"`
	Monitored         bool              `json:"monitored"`
	MonitorNewItems   string            `json:"monitorNewItems"`
	UseSceneNumbering bool              `json:"useSceneNumbering"`
	Runtime           int               `json:"runtime"`
	TvdbID            int               `json:"tvdbId"`
	TvRageID          int               `json:"tvRageId"`
	TvMazeID          int               `json:"tvMazeId"`
	FirstAired        time.Time         `json:"firstAired"`
	LastAired         time.Time         `json:"lastAired"`
	SeriesType        string            `json:"seriesType"`
	CleanTitle        string            `json:"cleanTitle"`
	ImdbID            string            `json:"imdbId"`
	TitleSlug         string            `json:"titleSlug"`
	RootFolderPath    string            `json:"rootFolderPath"`
	Folder            string            `json:"folder"`
	Certification     string            `json:"certification"`
	Genres            []string          `json:"genres"`
	Tags              []int             `json:"tags"`
	Added             time.Time         `json:"added"`
	AddOptions        AddOptions        `json:"addOptions"`
	Ratings           Ratings           `json:"ratings"`
	Statistics        Statistics        `json:"statistics"`
	EpisodesChanged   bool              `json:"episodesChanged"`
}

func (s *SeriesResource) Title() string       { return s.MainTitle }
func (s *SeriesResource) Description() string { return s.Overview }

func (s *SeriesResource) FilterValue() string { return s.MainTitle }

type AlternateTitles struct {
	Title             string `json:"title"`
	SeasonNumber      int    `json:"seasonNumber"`
	SceneSeasonNumber int    `json:"sceneSeasonNumber"`
	SceneOrigin       string `json:"sceneOrigin"`
	Comment           string `json:"comment"`
}
type Images struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}
type OriginalLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type Statistics struct {
	NextAiring        time.Time `json:"nextAiring"`
	PreviousAiring    time.Time `json:"previousAiring"`
	EpisodeFileCount  int       `json:"episodeFileCount"`
	EpisodeCount      int       `json:"episodeCount"`
	TotalEpisodeCount int       `json:"totalEpisodeCount"`
	SizeOnDisk        int       `json:"sizeOnDisk"`
	ReleaseGroups     []string  `json:"releaseGroups"`
	PercentOfEpisodes float64   `json:"percentOfEpisodes"`
}

type Seasons struct {
	SeasonNumber int              `json:"seasonNumber"`
	Monitored    bool             `json:"monitored"`
	Statistics   SeasonStatistics `json:"statistics"`
	Images       []Images         `json:"images"`
}

type AddOptions struct {
	IgnoreEpisodesWithFiles      bool   `json:"ignoreEpisodesWithFiles"`
	IgnoreEpisodesWithoutFiles   bool   `json:"ignoreEpisodesWithoutFiles"`
	Monitor                      string `json:"monitor"`
	SearchForMissingEpisodes     bool   `json:"searchForMissingEpisodes"`
	SearchForCutoffUnmetEpisodes bool   `json:"searchForCutoffUnmetEpisodes"`
}

type Ratings struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
}

type SeasonStatistics struct {
	SeasonCount       int      `json:"seasonCount"`
	EpisodeFileCount  int      `json:"episodeFileCount"`
	EpisodeCount      int      `json:"episodeCount"`
	TotalEpisodeCount int      `json:"totalEpisodeCount"`
	SizeOnDisk        int      `json:"sizeOnDisk"`
	ReleaseGroups     []string `json:"releaseGroups"`
	PercentOfEpisodes float64  `json:"percentOfEpisodes"`
}

type Client struct {
	URL    string
	APIKey string
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		URL:    url,
		APIKey: apiKey,
	}
}

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

func decode[T any](resp *http.Response) (T, error) {
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
