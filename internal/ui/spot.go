package ui

import (
	"strconv"

	. "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
	"github.com/ztimes2/surf-forecast/internal/meteo365surf"
)

// SpotPage returns a Node that renders the spot page.
func SpotPage(props SpotPageProps) Node {
	return HTML5(HTML5Props{
		Title:       props.Break.Name + " - Lighter surf forecasts",
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
				
				table {
					border-collapse: separate; 
					border-spacing: 10px 0px;
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
							Class("fs-3 fw-normal text-center mb-1"),
							Text(props.Break.Name),
						),
						H2(
							Class("fs-6 fw-light opacity-75 mb-3"),
							Text(props.Break.CountryName),
						),
						Div(
							mapIndex(props.ForecastIssue.Daily, func(i int, df *meteo365surf.DailyForecast) Node {
								return Group([]Node{
									H3(
										Class("fs-5 align-self-stretch mb-0 border-top pt-2 px-1"),
										Span(
											Class("fw-medium me-1"),
											Text(props.forecastWeekday(i)),
										),
										Span(
											Class("fw-light"),
											Text(props.forecastDate(i)),
										),
									),
									Div(
										Class("px-1 mb-4"),
										Style("margin: 0px -10px;"),
										Table(
											Class("table table-bordered"),
											THead(
												Tr(
													Th(
														Class("fw-light bg-transparent border-0 opacity-50"),
														Attr("scope", "col"),
													),
													Th(
														Class("fw-light bg-transparent border-0 opacity-50 text-center"),
														Attr("scope", "col"),
														Small(Text("Swell")),
													),
													Th(
														Class("fw-light bg-transparent border-0 opacity-50 text-center"),
														Attr("scope", "col"),
														Small(Text("Wind")),
													),
												),
											),
											TBody(
												mapIndex(df.Hourly, func(j int, hf meteo365surf.HourlyForecast) Node {
													return Tr(
														Th(
															Class("fw-light bg-transparent border-0 opacity-50 text-end py-3 px-0 text-nowrap"),
															Attr("scope", "row"),
															Small(Text(props.forecastHour(i, j))),
														),
														Td(
															Classes{
																"p-3 text-center": true,
																"border-bottom border-top rounded-top-3 rounded-bottom-3": len(df.Hourly) == 1,                                 // Only one hour is available
																"border-bottom border-top rounded-top-3":                  len(df.Hourly) > 1 && j == 0,                        // First hour among many
																"border-top-0 border-bottom rounded-bottom-3":             len(df.Hourly) > 1 && j == len(df.Hourly)-1,         // Last hour among many
																"border-top-0 border-bottom":                              len(df.Hourly) > 1 && j > 0 && j < len(df.Hourly)-1, // Hours in between many
															},
															Div(
																Class("row"),
																Div(
																	Class("col text-nowrap"),
																	Text(strconv.FormatFloat(hf.Swells.Primary.WaveHeightInMeters, 'f', -1, 64)),
																	Small(
																		Class("fw-light"),
																		Text(" m"),
																	),
																),
																Div(
																	Class("col text-nowrap"),
																	Text(strconv.FormatFloat(hf.Swells.Primary.PeriodInSeconds, 'f', -1, 64)),
																	Small(
																		Class("fw-light"),
																		Text(" s"),
																	),
																),
																Div(
																	Class("col text-nowrap"),
																	Text(strconv.FormatFloat(hf.WaveEnergyInKiloJoules, 'f', -1, 64)),
																	Small(
																		Class("fw-light"),
																		Text(" kJ"),
																	),
																),
															),
														),
														Td(
															Classes{
																"p-3 text-center": true,
																"border-bottom border-top rounded-top-3 rounded-bottom-3": len(df.Hourly) == 1,                                 // Only one hour is available
																"border-bottom border-top rounded-top-3":                  len(df.Hourly) > 1 && j == 0,                        // First hour among many
																"border-top-0 border-bottom rounded-bottom-3":             len(df.Hourly) > 1 && j == len(df.Hourly)-1,         // Last hour among many
																"border-top-0 border-bottom":                              len(df.Hourly) > 1 && j > 0 && j < len(df.Hourly)-1, // Hours in between many
															},
															Div(
																Class("row"),
																Div(
																	Class("col text-nowrap"),
																	Text(strconv.FormatFloat(hf.Wind.SpeedInKilometersPerHour, 'f', -1, 64)),
																	Small(
																		Class("fw-light"),
																		Text(" km/h"),
																	),
																),
																Div(
																	Class("col text-nowrap"),
																	Text(hf.Wind.State),
																),
															),
														),
													)
												})...,
											),
										),
									),
								})
							})...,
						),
					),
				),
				footer(),
			),
		},
	})
}

// SpotPageProps holds data needed for rendering the spot page.
type SpotPageProps struct {
	Break         meteo365surf.Break
	ForecastIssue *meteo365surf.ForecastIssue
}

// forecastWeekday returns a textual representation of a weekday by a daily forecast index.
func (p SpotPageProps) forecastWeekday(i int) string {
	if i == 0 {
		return "Today"
	}
	if i == 1 {
		return "Tomorrow"
	}
	return p.ForecastIssue.Daily[i].Timestamp.Format("Monday")
}

// forecastDate returns a textual representation of a date by a daily forecast index.
func (p SpotPageProps) forecastDate(i int) string {
	return p.ForecastIssue.Daily[i].Timestamp.Format("2 Jan")
}

// forecastHour returns a textual representation of an hour by indexes of daily and hourly forecasts respectively.
func (p SpotPageProps) forecastHour(i, j int) string {
	return p.ForecastIssue.Daily[i].Hourly[j].Timestamp.Format("3 pm")
}
