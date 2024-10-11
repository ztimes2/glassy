package ui

import (
	. "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// footer returns a Node that renders the footer section.
func footer() Node {
	return Footer(
		Class("row align-self-stretch"),
		Div(Class("col")),
		Div(
			Class("col col-10 col-md-12 col-lg-5 py-3"),
			P(
				Class("text-secondary fw-light text-center opacity-50 lh-sm"),
				Small(
					Text("The location and forecast data is obtained from "),
					A(
						Class("link-secondary text-decoration-none link-offset-1 fw-medium text-nowrap"),
						Href("https://www.surf-forecast.com"),
						Text("www.surf-forecast.com"),
					),
					Text(" via web scraping, it belongs to its original creators, and full credit is given to them."),
				),
			),
		),
		Div(Class("col")),
	)
}
