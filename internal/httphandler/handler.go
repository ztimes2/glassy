package httphandler

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ztimes2/glassy/internal/meteo365surf"
	"github.com/ztimes2/glassy/internal/ui"
)

// New initializes a new HTTP handler configured to serve the application's requests.
func New(scraper *meteo365surf.Scraper) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleIndex())
	mux.HandleFunc("GET /search", handleSearch(scraper))
	mux.HandleFunc("GET /breaks/{break_id}/forecasts/latest", handleLatestForecast(scraper))

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
		var (
			breaks []meteo365surf.BreakSearchResult
			err    error
		)

		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query != "" {
			breaks, err = scraper.SearchBreaks(query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		page := ui.SearchPage(ui.SearchPageProps{
			SearchQuery: query,
			Breaks:      breaks,
		})

		buf := new(bytes.Buffer)
		if err := page.Render(buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cacheResponse(w, time.Hour)
		_, _ = w.Write(buf.Bytes())
	}
}

func handleLatestForecast(scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.TrimSpace(r.PathValue("break_id")))
		if err != nil {
			http.Error(w, "invalid break id", http.StatusBadRequest)
			return
		}

		brk, err := scraper.Break(id)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		iss, err := scraper.LatestForecastIssue(brk.Slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		page := ui.LatestForecastPage(ui.LatestForecastPageProps{
			Break:         brk,
			ForecastIssue: iss,
		})

		buf := new(bytes.Buffer)
		if err := page.Render(buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cacheResponse(w, time.Hour)
		_, _ = w.Write(buf.Bytes())
	}
}

func cacheResponse(w http.ResponseWriter, d time.Duration) {
	age := strconv.Itoa(int(d.Seconds()))
	w.Header().Set("Cache-Control", "max-age="+age)
}
