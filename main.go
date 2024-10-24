package main

import (
	"net/http"

	"github.com/ztimes2/glassy/internal/meteo365"
	"github.com/ztimes2/glassy/internal/router"
)

func main() {
	scraper := meteo365.NewScraper()

	r := router.New(scraper)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
