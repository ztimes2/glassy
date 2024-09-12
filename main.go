package main

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/ztimes2/surf-forecast/internal/httphandler"
)

//go:embed templates
var fs embed.FS

func main() {
	tpl, err := template.ParseFS(fs, "templates/*.html")
	if err != nil {
		panic(err)
	}

	h := httphandler.New(tpl)

	err = http.ListenAndServe(":8080", h)
	if err != nil {
		panic(err)
	}
}
