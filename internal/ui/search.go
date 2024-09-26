package ui

import (
	"strconv"

	. "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
)

// SearchPage returns a Node that renders the search page.
func SearchPage(props SearchPageProps) Node {
	return HTML5(HTML5Props{
		Title:       "Lighter surf forecasts",
		Description: "It's like www.surf-forecast.com but lighter.",
		Head: []Node{
			Link(
				Href("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css"),
				Rel("stylesheet"),
				Integrity("sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH"),
				CrossOrigin("anonymous"),
			),
			Script(Src("https://unpkg.com/htmx.org@2.0.2")),
			StyleEl(Raw(`
				/* Modify Bootstrap's vh-100 class to properly support iOS screens. */
				@supports (-webkit-touch-callout: none) {
					.vh-100 {
						height: -webkit-fill-available !important;
					}
				}

				#search-results > * .list-group-item {
					background-color: transparent !important;
				}
				#search-results > * .list-group-item:hover {
					background-color: var(--bs-gray-200) !important;
				}
			`)),
		},
		Body: []Node{
			Class("container bg-light"),
			Div(
				Class("d-flex flex-column justify-content-between align-items-center vh-100 gap-4"),
				Header(),
				Main(
					Class("d-flex flex-column justify-content-center align-items-center align-self-stretch gap-2"),
					Div(
						Class("d-flex flex-column justify-content-center align-items-center"),
						H1(
							Class("fs-3 fw-light text-center"),
							Text("It's like "),
							A(
								Class("link-primary link-offset-1"),
								Href("https://www.surf-forecast.com"),
								Text("surf-forecast.com"),
							),
							Text(" but "),
							Span(
								Class("fw-semibold fst-italic"),
								Text("lighter"),
							),
						),
					),
					Div(
						Class("row align-self-stretch"),
						Div(Class("col")),
						Div(
							Class("col col-12 col-md-8 col-lg-5"),
							Input(
								ID("search-bar"),
								Class("form-control form-control-lg fw-light"),
								Type("search"),
								Placeholder("Start typing to find surf spots"),
								Value(props.SearchQuery),
								hx.Get("/search"),
								Name("q"),
								hx.Select("#search-results"),
								hx.Trigger("input changed"),
								hx.Target("#search-results"),
								hx.Swap("outerHTML"),
								hx.ReplaceURL("true"),
							),
						),
						Div(Class("col")),
					),
					Div(
						ID("search-results"),
						Class("row align-self-stretch"),
						If(
							len(props.Breaks) > 0,
							Group([]Node{
								Div(Class("col")),
								Div(
									Class("col col-12 col-md-8 col-lg-5 px-3 pt-2 list-group list-group-flush"),
									Group(Map(props.Breaks, func(b meteo365surf.BreakSearchResult) Node {
										return A(
											Class("list-group-item list-group-item-action py-2"),
											Href("/breaks/"+strconv.Itoa(b.ID)+"/forecasts/latest"),
											H6(
												Class("mb-0 fs-6"),
												Text(b.Name),
											),
											Small(
												Class("opacity-75"),
												Text(b.CountryName),
											),
										)
									})),
								),
								Div(Class("col")),
							}),
						),
					),
				),
				footer(),
			),
			Script(
				If(
					props.SearchQuery == "",
					Raw(`document.getElementById("search-bar").focus();`),
				),
			),
		},
	})
}

// SearchPageProps holds data needed for rendering the search page.
type SearchPageProps struct {
	SearchQuery string
	Breaks      []meteo365surf.BreakSearchResult
}
