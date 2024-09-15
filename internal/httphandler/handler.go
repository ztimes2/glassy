package httphandler

import (
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
		var data struct {
			SearchQuery string
			Breaks      []meteo365surf.Break
		}

		urlQuery := r.URL.Query()

		data.SearchQuery = strings.TrimSpace(urlQuery.Get("q"))
		if data.SearchQuery != "" {
			breaks, err := scraper.SearchBreaks(data.SearchQuery)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data.Breaks = breaks
		}

		err := tpl.ExecuteTemplate(w, "search.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO add caching
	}
}

func handleSpot(tpl *template.Template, scraper *meteo365surf.Scraper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		b, err := scraper.Break(slug)
		if err != nil {
			if errors.Is(err, meteo365surf.ErrBreakNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tpl.ExecuteTemplate(w, "spot.html", b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO add caching
	}
}
