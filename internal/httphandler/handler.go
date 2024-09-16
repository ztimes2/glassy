package httphandler

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
)

// New initializes a new HTTP handler configured to serve the application's requests.
func New(tpl *template.Template, scraper *meteo365surf.Scraper) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleIndex())
	mux.HandleFunc("GET /search", handleSearch(tpl, scraper))
	mux.HandleFunc("GET /spots/{name}", handleSpot(tpl, scraper))

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

func handleSearch(tpl *template.Template, scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var templateData struct {
			SearchQuery string
			Breaks      []meteo365surf.Break
		}

		urlQuery := r.URL.Query()

		templateData.SearchQuery = strings.TrimSpace(urlQuery.Get("q"))
		if templateData.SearchQuery != "" {
			breaks, err := scraper.SearchBreaks(templateData.SearchQuery)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			templateData.Breaks = breaks
		}

		buf := new(bytes.Buffer)

		err := tpl.ExecuteTemplate(buf, "search.html", templateData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(buf.Bytes())

		// TODO add caching
	}
}

func handleSpot(tpl *template.Template, scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var templateData struct {
			Break    meteo365surf.Break
			Forecast *meteo365surf.Forecast
		}

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

		templateData.Break, err = scraper.Break(slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateData.Forecast, err = scraper.LatestForecast(slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf := new(bytes.Buffer)

		err = tpl.ExecuteTemplate(buf, "spot.html", templateData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(buf.Bytes())

		// TODO add caching
	}
}
