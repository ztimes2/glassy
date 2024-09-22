package httphandler

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
	"github.com/ztimes2/surf-forecast/internal/ui"
)

// New initializes a new HTTP handler configured to serve the application's requests.
func New(scraper *meteo365surf.Scraper) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleIndex())
	mux.HandleFunc("GET /search", handleSearch(scraper))
	mux.HandleFunc("GET /spots/{name}", handleSpot(scraper))

	return mux
}

func handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The "/" pattern matches everything, so we need to check that we're at the root here.
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, "/search", http.StatusMovedPermanently)
	}
}

func handleSearch(scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var props ui.SearchPageProps

		urlQuery := r.URL.Query()

		props.SearchQuery = strings.TrimSpace(urlQuery.Get("q"))
		if props.SearchQuery != "" {
			breaks, err := scraper.SearchBreaks(props.SearchQuery)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			props.Breaks = breaks
		}

		buf := new(bytes.Buffer)

		err := ui.SearchPage(props).Render(buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(buf.Bytes())

		// TODO add caching
	}
}

func handleSpot(scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var props ui.SpotPageProps

		name := strings.TrimSpace(r.PathValue("name"))

		slug, err := scraper.BreakSlug(name)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		props.Break, err = scraper.Break(slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		props.ForecastIssue, err = scraper.LatestForecastIssue(slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf := new(bytes.Buffer)

		err = ui.SpotPage(props).Render(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(buf.Bytes())

		// TODO add caching
	}
}
