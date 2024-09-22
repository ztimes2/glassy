package main

import (
	"net/http"

	"github.com/ztimes2/surf-forecast/internal/httphandler"
	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
)

func main() {
	scraper := meteo365surf.NewScraper()

	h := httphandler.New(scraper)

	err := http.ListenAndServe(":8080", h)
	if err != nil {
		panic(err)
	}
}
