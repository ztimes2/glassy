package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/ztimes2/glassy/internal/meteo365"
	"github.com/ztimes2/glassy/internal/router"
)

//go:embed all:static
var static embed.FS

func main() {
	scraper := meteo365.NewScraper()

	assets, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	r := router.New(scraper, assets)

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
