package meteo365

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tkuchiki/go-timezone"
	"github.com/ztimes2/glassy/internal/htmlutil"
	"golang.org/x/net/html"
)

// LatestForecastIssue returns latest forecast issue for a surf break by its slug for 8 or 9
// subsequent days. The returned forecast's timestamps use the surf break's local timezone.
// It returns ErrBreakNotFound for non-existent surf breaks.
func (s *Scraper) LatestForecastIssue(slug string) (*ForecastIssue, error) {
	path := "/breaks/" + slug + "/forecasts/latest"

	req, err := http.NewRequest(http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("could not prepare request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrBreakNotFound
		}
		return nil, fmt.Errorf("received response with %d status code", resp.StatusCode)
	}

	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse response body as html: %w", err)
	}

	forecast, err := scrapeForecast(node, s.timezone)
	if err != nil {
		return nil, fmt.Errorf("could not scrape html: %w", err)
	}

	return forecast, nil
}

// ForecastIssue holds a forecast issue for multiple days.
type ForecastIssue struct {
	// IssuedAt holds a timestamp of when the given forecast was issued by www.surf-forecast.com
	// using the surf break's local timezone.
	IssuedAt time.Time
	Daily    []*DailyForecast
}

// newForecastIssue combines the scraped forecast data into ForecastIssue.
func newForecastIssue(
	issuedAt time.Time,
	days []int,
	hours [][]int,
	ratings [][]int,
	swells [][]Swells,
	waveEnergies [][]float64,
	winds [][]wind,
	windStates [][]string,
) (*ForecastIssue, error) {

	if len(days) != len(hours) {
		return nil, errors.New("days and hours must have equal number of elements")
	}
	if len(days) != len(ratings) {
		return nil, errors.New("days and ratings must have equal number of elements")
	}
	if len(days) != len(swells) {
		return nil, errors.New("days and swells must have equal number of elements")
	}
	if len(days) != len(waveEnergies) {
		return nil, errors.New("days and wave energies must have equal number of elements")
	}
	if len(days) != len(winds) {
		return nil, errors.New("days and winds must have equal number of elements")
	}
	if len(days) != len(windStates) {
		return nil, errors.New("days and wind states must have equal number of elements")
	}

	var (
		forecasts = make([]*DailyForecast, len(days))
		year      = issuedAt.Year()
		month     = issuedAt.Month()

		previous *DailyForecast
	)
	for i := range forecasts {
		if previous != nil {
			// Handle the case when a forecast contains days of two subsequent months.
			if previous.Timestamp.Day() > days[i] {
				if month+1 > time.December {
					month = time.January
				}
				month++
			}

			// Handle the case when a forecast contains days of two subsequent years.
			if previous.Timestamp.Month() > month {
				year++
			}
		}

		f, err := newDailyForecast(
			issuedAt.Location(),
			issuedAt.Year(),
			month,
			days[i],
			hours[i],
			ratings[i],
			swells[i],
			waveEnergies[i],
			winds[i],
			windStates[i],
		)
		if err != nil {
			return nil, fmt.Errorf("could not create forecast: %w", err)
		}

		forecasts[i] = f
		previous = f
	}

	return &ForecastIssue{
		IssuedAt: issuedAt,
		Daily:    forecasts,
	}, nil
}

// DailyForecast holds a forecast for a single day broken down into hours.
type DailyForecast struct {
	// Timestamp holds a date of the day the underlying hourly forecasts belong to
	// using the surf break's local timezone.
	Timestamp time.Time
	Hourly    []HourlyForecast
}

// newDailyForecast combines the scraped forecast data of a single day into DailyForecast.
func newDailyForecast(
	l *time.Location,
	year int,
	month time.Month,
	day int,
	hours []int,
	ratings []int,
	swells []Swells,
	waveEnergies []float64,
	winds []wind,
	windStates []string,
) (*DailyForecast, error) {

	if len(hours) != len(ratings) {
		return nil, errors.New("hours and ratings must have equal number of elements")
	}
	if len(hours) != len(swells) {
		return nil, errors.New("hours and swells must have equal number of elements")
	}
	if len(hours) != len(waveEnergies) {
		return nil, errors.New("hours and wave energies must have equal number of elements")
	}
	if len(hours) != len(winds) {
		return nil, errors.New("hours and winds must have equal number of elements")
	}
	if len(hours) != len(windStates) {
		return nil, errors.New("hours and wind states must have equal number of elements")
	}

	forecasts := make([]HourlyForecast, len(hours))
	for i := range forecasts {
		forecasts[i].Timestamp = time.Date(year, month, day, hours[i], 0, 0, 0, l)
		forecasts[i].Rating = ratings[i]
		forecasts[i].Swells = swells[i]
		forecasts[i].WaveEnergyInKiloJoules = waveEnergies[i]
		forecasts[i].Wind = Wind{
			SpeedInKilometersPerHour:     winds[i].speed,
			DirectionToInDegrees:         winds[i].degrees,
			DirectionFromInCompassPoints: winds[i].letters,
			State:                        windStates[i],
		}
	}

	return &DailyForecast{
		Timestamp: time.Date(year, month, day, 0, 0, 0, 0, l),
		Hourly:    forecasts,
	}, nil
}

// HourlyForecast holds a forecast for a single hour.
type HourlyForecast struct {
	// Timestamp holds a timestamp of the given forecast's day and hour.
	Timestamp time.Time

	// Rating holds a rating score ranging from 0 to 10 that represents the surf
	// quality according to www.surf-forecast.com.
	Rating                 int
	Swells                 Swells
	WaveEnergyInKiloJoules float64
	Wind                   Wind
}

// Swells holds information about primary and secondary swells.
type Swells struct {
	Primary   Swell
	Secondary []Swell
}

// Swell holds information about a swell.
type Swell struct {
	PeriodInSeconds              float64
	DirectionToInDegrees         float64
	DirectionFromInCompassPoints string
	WaveHeightInMeters           float64
}

// Wind holds information about a wind.
type Wind struct {
	SpeedInKilometersPerHour     float64
	DirectionToInDegrees         float64
	DirectionFromInCompassPoints string
	State                        string
}

func scrapeForecast(n *html.Node, tz *timezone.Timezone) (*ForecastIssue, error) {
	issuedAt, err := scrapeIssueTimestamp(n, tz)
	if err != nil {
		return nil, fmt.Errorf("could not scrape issue date: %w", err)
	}

	tableNode, ok := htmlutil.FindOne(n, htmlutil.WithClassEqual("forecast-table__basic"))
	if !ok {
		return nil, errors.New("could not find table node")
	}

	days, err := scrapeDays(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape days: %w", err)
	}

	hours, err := scrapeHours(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape hours: %w", err)
	}

	ratings, err := scrapeRatings(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape ratings: %w", err)
	}

	swells, err := scrapeSwells(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape swells: %w", err)
	}

	waveEnergies, err := scrapeWaveEnergies(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape wave energies: %w", err)
	}

	winds, err := scrapeWinds(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape winds: %w", err)
	}

	windStates, err := scrapeWindStates(tableNode)
	if err != nil {
		return nil, fmt.Errorf("could not scrape wind states: %w", err)
	}

	return newForecastIssue(
		issuedAt,
		days,
		hours,
		ratings,
		swells,
		waveEnergies,
		winds,
		windStates,
	)
}

func scrapeIssueTimestamp(n *html.Node, tz *timezone.Timezone) (time.Time, error) {
	issueNode, ok := htmlutil.FindOne(n, htmlutil.WithClassEqual("break-header-dynamic__issued"))
	if !ok {
		return time.Time{}, errors.New("could not find issue node")
	}

	issueTextNode := issueNode.LastChild
	if issueTextNode == nil {
		return time.Time{}, errors.New("could not find issue text node")
	}

	parts := strings.Split(issueTextNode.Data, " ")
	if len(parts) != 12 {
		return time.Time{}, fmt.Errorf("unexpected issue text: %q", issueTextNode.Data)
	}

	hourText := parts[5]
	clockPeriodText := parts[6]
	dayText := parts[8]
	monthText := parts[9]
	yearText := parts[10]
	tzAbbr := parts[11]

	hour, err := parseTwelveClockHour(hourText)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse issue hour: %w", err)
	}

	clockPeriod, err := parseClockPeriod(clockPeriodText)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse clock period: %w", err)
	}

	hour = toTwentyFourClockHour(hour, clockPeriod)

	day, err := parseDay(dayText)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse issue day: %w", err)
	}

	month, err := parseMonthShort(monthText)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse issue month: %w", err)
	}

	year, err := strconv.Atoi(yearText)
	if err != nil {
		return time.Time{}, fmt.Errorf("issue year not integer: %q", yearText)
	}

	// github.com/tkuchiki/go-timezone package is not able to parse timezone offsets (i.e. +06, -05, etc.).
	// Therefore, the time package is utilized for such cases instead.
	if strings.HasPrefix(tzAbbr, "+") || strings.HasPrefix(tzAbbr, "-") {
		t, err := time.Parse("-07", tzAbbr)
		if err != nil {
			return time.Time{}, fmt.Errorf("could not parse time location for %q", tzAbbr)
		}

		return time.Date(year, month, day, hour, 0, 0, 0, t.Location()), nil
	}

	timezones, err := tz.GetTimezones(tzAbbr)
	if err != nil || len(timezones) == 0 {
		return time.Time{}, fmt.Errorf("could not find timezones for %q abbreviation: %w", tzAbbr, err)
	}

	loc, err := time.LoadLocation(timezones[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("could not find time location for %q", timezones[0])
	}

	return time.Date(year, month, day, hour, 0, 0, 0, loc), nil
}

func parseDay(s string) (int, error) {
	day, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not integer: %q", s)
	}

	if day < 0 || day > 31 {
		return 0, fmt.Errorf("not month day: %q", s)
	}

	return day, nil
}

func parseMonthShort(s string) (time.Month, error) {
	switch s {
	case "Jan":
		return time.January, nil
	case "Feb":
		return time.February, nil
	case "Mar":
		return time.March, nil
	case "Apr":
		return time.April, nil
	case "May":
		return time.May, nil
	case "Jun":
		return time.June, nil
	case "Jul":
		return time.July, nil
	case "Aug":
		return time.August, nil
	case "Sep":
		return time.September, nil
	case "Oct":
		return time.October, nil
	case "Nov":
		return time.November, nil
	case "Dec":
		return time.December, nil
	default:
		return time.Month(0), fmt.Errorf("invalid short month: %q", s)
	}
}

func scrapeDays(n *html.Node) ([]int, error) {
	daysNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassContaining("forecast-table__row", "forecast-table-days"),
		htmlutil.WithAttributeEqual("data-row-name", "days"),
	)
	if !ok {
		return nil, errors.New("could not find days node")
	}

	var days []int
	if err := htmlutil.ForEach(daysNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			day, err := scrapeDay(n)
			if err != nil {
				return fmt.Errorf("could not scrape day: %w", err)
			}

			days = append(days, day)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return days, nil
}

func scrapeDay(n *html.Node) (int, error) {
	dayNameAttr, ok := htmlutil.Attribute(n, "data-day-name")
	if !ok {
		return 0, errors.New("could not find day name attribute")
	}

	t, err := time.Parse("Mon_02", dayNameAttr.Val)
	if err != nil {
		return 0, fmt.Errorf("could not parse day name attribute: %q", dayNameAttr.Val)
	}

	return t.Day(), nil
}

func scrapeHours(n *html.Node) ([][]int, error) {
	hoursNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassContaining("forecast-table__row", "forecast-table-time"),
		htmlutil.WithAttributeEqual("data-row-name", "time"),
	)
	if !ok {
		return nil, errors.New("could not find hours node")
	}

	var (
		allHours [][]int
		hours    []int
	)
	if err := htmlutil.ForEach(hoursNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			hour, err := scrapeHour(n)
			if err != nil {
				return fmt.Errorf("could not scrape hour: %w", err)
			}

			hours = append(hours, hour)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allHours = append(allHours, hours)
				hours = []int{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allHours, nil
}

func scrapeHour(n *html.Node) (int, error) {
	nodes := htmlutil.Find(n, htmlutil.WithClassEqual("forecast-table__value"))
	if len(nodes) != 2 {
		return 0, errors.New("unexpected table values")
	}

	hourTextNode := nodes[0].FirstChild
	if hourTextNode == nil {
		return 0, errors.New("could not find hour text node")
	}

	hour, err := parseTwelveClockHour(hourTextNode.Data)
	if err != nil {
		return 0, fmt.Errorf("could not parse hour: %w", err)
	}

	periodTextNode := nodes[1].FirstChild
	if periodTextNode == nil {
		return 0, errors.New("could not find clock period text node")
	}

	period, err := parseClockPeriod(periodTextNode.Data)
	if err != nil {
		return 0, fmt.Errorf("could not parse clock period: %w", err)
	}

	return toTwentyFourClockHour(hour, period), nil
}

func parseTwelveClockHour(s string) (int, error) {
	hour, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not integer: %q", s)
	}

	if hour < 0 || hour > 12 {
		return 0, fmt.Errorf("not 12 clock hour: %q", s)
	}

	// This is needed to map 0AM to 12AM correctly.
	if hour == 0 {
		return 12, nil
	}

	return hour, nil
}

type clockPeriod int

const (
	beforeMidday clockPeriod = iota
	afterMidday
)

func parseClockPeriod(s string) (clockPeriod, error) {
	switch strings.ToUpper(s) {
	case "AM":
		return beforeMidday, nil
	case "PM":
		return afterMidday, nil
	default:
		return clockPeriod(0), fmt.Errorf("invalid clock period: %q", s)
	}
}

func toTwentyFourClockHour(hour int, p clockPeriod) int {
	if p == beforeMidday {
		if hour == 12 {
			return 0
		}
		return hour
	}
	if hour == 12 {
		return hour
	}
	return hour + 12
}

func scrapeRatings(n *html.Node) ([][]int, error) {
	ratingsNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassContaining("forecast-table__row", "forecast-table-rating"),
		htmlutil.WithAttributeEqual("data-row-name", "rating"),
	)
	if !ok {
		return nil, errors.New("could not find ratings node")
	}

	var (
		allRatings [][]int
		ratings    []int
	)
	if err := htmlutil.ForEach(ratingsNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			ratingNode, ok := htmlutil.FindOne(n, htmlutil.WithClassContaining("star-rating__rating"))
			if !ok {
				return errors.New("could not find rating node")
			}

			ratingTextNode := ratingNode.FirstChild
			if ratingTextNode == nil {
				return errors.New("could not find wave energy text node")
			}

			rating, err := parseRating(ratingTextNode.Data)
			if err != nil {
				return fmt.Errorf("could not parse rating: %w", err)
			}

			ratings = append(ratings, rating)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allRatings = append(allRatings, ratings)
				ratings = []int{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allRatings, nil
}

func parseRating(s string) (int, error) {
	if s == "!" {
		// For some forecasts the rating can be represented as "!" supposedly indicating
		// rough conditions. Since we are using numerical rating representation, let's just
		// use 11 as a special case.
		return 11, nil
	}

	rating, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not integer: %q", s)
	}

	if rating < 0 || rating > 10 {
		return 0, fmt.Errorf("invalid rating: %q", s)
	}

	return rating, nil
}

func scrapeSwells(n *html.Node) ([][]Swells, error) {
	swellsNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassEqual("forecast-table__row"),
		htmlutil.WithAttributeEqual("data-row-name", "wave-height"),
	)
	if !ok {
		return nil, errors.New("could not find swells node")
	}

	var (
		allSwells [][]Swells
		swells    []Swells
	)
	if err := htmlutil.ForEach(swellsNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			hourlySwells, err := scrapeHourlySwells(n)
			if err != nil {
				return fmt.Errorf("could not scrape hourly swells: %w", err)
			}

			var s Swells
			if len(hourlySwells) > 0 {
				s = Swells{
					Primary:   hourlySwells[0],
					Secondary: hourlySwells[1:],
				}
			}
			swells = append(swells, s)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allSwells = append(allSwells, swells)
				swells = []Swells{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allSwells, nil
}

func scrapeHourlySwells(n *html.Node) ([]Swell, error) {
	attr, ok := htmlutil.Attribute(n, "data-swell-state")
	if !ok {
		return nil, errors.New("could not find swells attribute")
	}

	swells, err := unmarshalSwells([]byte(attr.Val))
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal swells: %w", err)
	}

	return swells, nil
}

func unmarshalSwells(b []byte) ([]Swell, error) {
	var payload []*swell
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, fmt.Errorf("could not unmarshal payload: %w", err)
	}

	var swells []Swell
	for _, p := range payload {
		if p == nil {
			continue
		}

		swells = append(swells, Swell{
			PeriodInSeconds:              p.Period,
			DirectionToInDegrees:         p.Angle,
			DirectionFromInCompassPoints: p.Letters,
			WaveHeightInMeters:           p.Height,
		})
	}

	return swells, nil
}

type swell struct {
	Period  float64 `json:"period"`
	Angle   float64 `json:"angle"`
	Letters string  `json:"letters"`
	Height  float64 `json:"height"`
}

func scrapeWaveEnergies(n *html.Node) ([][]float64, error) {
	energiesNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassEqual("forecast-table__row"),
		htmlutil.WithAttributeEqual("data-row-name", "energy"),
	)
	if !ok {
		return nil, errors.New("could not find wave energies node")
	}

	var (
		allEnergies [][]float64
		energies    []float64
	)
	if err := htmlutil.ForEach(energiesNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			energy, err := scrapeWaveEnergy(n)
			if err != nil {
				return fmt.Errorf("could not scrape wave energy: %w", err)
			}

			energies = append(energies, energy)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allEnergies = append(allEnergies, energies)
				energies = []float64{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allEnergies, nil
}

func scrapeWaveEnergy(n *html.Node) (float64, error) {
	energyNode := n.FirstChild
	if energyNode == nil {
		return 0, errors.New("could not find wave energy node")
	}

	energyTextNode := energyNode.FirstChild
	if energyTextNode == nil {
		return 0, errors.New("could not find wave energy text node")
	}

	energy, err := parseWaveEnergy(energyTextNode.Data)
	if err != nil {
		return 0, fmt.Errorf("could not parse wave energy: %w", err)
	}

	return energy, nil
}

func parseWaveEnergy(s string) (float64, error) {
	energy, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("not float: %q", s)
	}

	if energy < 0 {
		return 0, fmt.Errorf("invalid wave energy: %q", s)
	}

	return energy, nil
}

func scrapeWinds(n *html.Node) ([][]wind, error) {
	windsNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassEqual("forecast-table__row"),
		htmlutil.WithAttributeEqual("data-row-name", "wind"),
	)
	if !ok {
		return nil, errors.New("could not find winds node")
	}

	var (
		allWinds [][]wind
		winds    []wind
	)
	if err := htmlutil.ForEach(windsNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			w, err := scrapeWind(n)
			if err != nil {
				return fmt.Errorf("could not scrape wind: %w", err)
			}

			winds = append(winds, w)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allWinds = append(allWinds, winds)
				winds = []wind{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allWinds, nil
}

func scrapeWind(n *html.Node) (wind, error) {
	iconNode, ok := htmlutil.FindOne(n, htmlutil.WithClassEqual("wind-icon"))
	if !ok {
		return wind{}, errors.New("could not find wind icon node")
	}

	speedAttr, ok := htmlutil.Attribute(iconNode, "data-speed")
	if !ok {
		return wind{}, errors.New("could not find wind speed attribute")
	}

	speed, err := parseWindSpeed(speedAttr.Val)
	if err != nil {
		return wind{}, fmt.Errorf("could not parse wind speed: %w", err)
	}

	degrees, err := scrapeWindDirectionDegrees(iconNode)
	if err != nil {
		return wind{}, fmt.Errorf("could not scrape wind direction degrees: %w", err)
	}

	lettersNode, ok := htmlutil.FindOne(iconNode, htmlutil.WithClassEqual("wind-icon__letters"))
	if !ok {
		return wind{}, errors.New("could not find wind direction letters node")
	}

	lettersTextNode := lettersNode.FirstChild
	if lettersTextNode == nil {
		return wind{}, errors.New("could not find wind direction letters text node")
	}

	return wind{
		speed:   speed,
		degrees: degrees,
		letters: lettersTextNode.Data,
	}, nil
}

type wind struct {
	speed   float64
	degrees float64
	letters string
}

func scrapeWindDirectionDegrees(n *html.Node) (float64, error) {
	arrowNode, ok := htmlutil.FindOne(n, htmlutil.WithClassEqual("wind-icon__arrow"))
	if !ok {
		return 0, errors.New("could not find wind direction arrow node")
	}

	attr, ok := htmlutil.Attribute(arrowNode, htmlutil.AttributeTransform)
	if !ok {
		return 0, errors.New("could not find transform attribute")
	}

	degreesText := strings.TrimPrefix(attr.Val, "rotate(")
	degreesText = strings.TrimSuffix(degreesText, ")")

	degrees, err := parseWindDirectionDegrees(degreesText)
	if err != nil {
		return 0, fmt.Errorf("could not parse wind direction degrees: %w", err)
	}

	return degrees, nil
}

func parseWindDirectionDegrees(s string) (float64, error) {
	degrees, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("not float: %q", s)
	}

	if degrees < 0 || degrees > 360 {
		return 0, fmt.Errorf("invalid wind direction degrees: %q", s)
	}

	return degrees, nil
}

func parseWindSpeed(s string) (float64, error) {
	speed, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("not float: %q", s)
	}

	if speed < 0 {
		return 0, fmt.Errorf("invalid wind speed: %q", s)
	}

	return speed, nil
}

func scrapeWindStates(n *html.Node) ([][]string, error) {
	statesNode, ok := htmlutil.FindOne(
		n,
		htmlutil.WithClassEqual("forecast-table__row"),
		htmlutil.WithAttributeEqual("data-row-name", "wind-state"),
	)
	if !ok {
		return nil, errors.New("could not find wind states node")
	}

	var (
		allStates [][]string
		states    []string
	)
	if err := htmlutil.ForEach(statesNode, func(n *html.Node) error {
		if htmlutil.ClassContains(n, "forecast-table__cell") {
			state, err := scrapeWindState(n)
			if err != nil {
				return fmt.Errorf("could not scrape wind state: %w", err)
			}

			states = append(states, state)

			isDayEnd := htmlutil.ClassContains(n, "is-day-end")
			if isDayEnd {
				allStates = append(allStates, states)
				states = []string{}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return allStates, nil
}

func scrapeWindState(n *html.Node) (string, error) {
	var ss []string
	htmlutil.ForEach(n, func(n *html.Node) error {
		if n.Type == html.TextNode {
			ss = append(ss, n.Data)
		}
		return nil
	})

	state := strings.Join(ss, "")
	if state == "" {
		return "", errors.New("invalid wind state")
	}

	return state, nil
}
