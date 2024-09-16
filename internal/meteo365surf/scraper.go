package meteo365surf

import (
	"net/http"
	"time"

	"github.com/tkuchiki/go-timezone"
)

const (
	defaultBaseURL = "https://www.surf-forecast.com"
	defaultTimeout = 10 * time.Second
)

// Scraper is a web scraper that sends requests to www.surf-forecast.com and scrapes
// data from its responses.
type Scraper struct {
	baseURL  string
	client   *http.Client
	timezone *timezone.Timezone
}

// NewScraper initializes a new Scraper.
func NewScraper() *Scraper {
	return &Scraper{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// This prevents from automatically following redirects because the BreakSlug method
				// relies on redirect response which need to be intercepted.
				return http.ErrUseLastResponse
			},
		},
		timezone: timezone.New(),
	}
}
