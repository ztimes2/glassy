package httphandler

import (
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
	}
}
