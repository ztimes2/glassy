package httphandler

import (
	"html/template"
	"net/http"
)

// New initializes a new HTTP handler configured to serve the application's requests.
func New(t *template.Template) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleIndex(t))

	return mux
}

func handleIndex(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The "/" pattern matches everything, so we need to check that we're at the root here.
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		err := tpl.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
