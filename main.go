package main

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/ztimes2/surf-forecast/internal/httphandler"
	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
)

//go:embed templates
var fs embed.FS

func main() {
	tpl, err := template.New("").
		Funcs(template.FuncMap{
			"sub": func(a, b int) int {
				return a - b
			},
		}).
		ParseFS(fs, "templates/*.html")
	if err != nil {
		panic(err)
	}

	scraper := meteo365surf.NewScraper()

	h := httphandler.New(tpl, scraper)

	err = http.ListenAndServe(":8080", h)
	if err != nil {
		panic(err)
	}
}
