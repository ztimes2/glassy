package meteo365surf

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://www.surf-forecast.com"
	defaultTimeout = 10 * time.Second
)

// Scraper is a web scraper that sends requests to www.surf-forecast.com and scrapes
// data from its responses.
type Scraper struct {
	baseURL string
	client  *http.Client
}

// NewScraper initializes a new Scraper.
func NewScraper() *Scraper {
	return &Scraper{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}
