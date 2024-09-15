package meteo365surf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ztimes2/surf-forecast/internal/htmlutil"
	"golang.org/x/net/html"
)

var (
	// ErrBreakNotFound indicates that a surf break could not be found.
	ErrBreakNotFound = errors.New("surf break not found")
)

// SearchBreaks searches for surf breaks using a text query.
func (s *Scraper) SearchBreaks(query string) ([]Break, error) {
	u, err := url.Parse(s.baseURL + "/breaks/ac_location_name")
	if err != nil {
		return nil, fmt.Errorf("could not prepare request url: %w", err)
	}

	u.RawQuery = url.Values{
		"query": []string{query},
	}.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not prepare request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received response with %d status code", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	// The search response's payload contains a 2D JSON-alike array of strings
	// that uses single quotes to represent a string.
	//
	// Example: [['a','b','c'],['a','b','c']]
	//
	// Therefore, these single quotes need to be replaced with double quotes in
	// order to make JSON unmarshaling work properly.
	body = bytes.ReplaceAll(body, []byte(`'`), []byte(`"`))

	var results [][]string
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	var breaks []Break
	for _, result := range results {
		if len(result) != 3 {
			return nil, fmt.Errorf("unexpected search result: %q", result)
		}

		// Each search result can represent either a surf break, a region, a country, or
		// some other type of locality.
		//
		// The first element is an ID which can be used to distinguish a result's type.
		// IDs of surf breaks are numerical values (i.e. "123", "456", etc.) and other types
		// contain special prefixes like "re" for regions (i.e. "re123", "re456", etc.),
		// "co" for countries (i.e. "co123", "co456", etc.), and so on.
		//
		// Therefore, let's ignore results that have non-numerical IDs since we are only
		// interested in returning surf breaks.
		if _, err := strconv.Atoi(result[0]); err != nil {
			continue
		}

		breaks = append(breaks, Break{
			Name:        result[1],
			CountryName: result[2],
		})
	}

	return breaks, nil
}

// Break holds information about a surf break.
type Break struct {
	Name        string
	CountryName string
}

// BreakSlug returns a surf break's slug by its name. It returns ErrBreakNotFound for non-existent surf breaks.
func (s *Scraper) BreakSlug(name string) (string, error) {
	resp, err := s.client.PostForm(s.baseURL+"/breaks/catch", url.Values{
		"query": []string{name},
	})
	if err != nil {
		return "", fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("received response with %d status code", resp.StatusCode)
	}

	redirectURL, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		return "", fmt.Errorf("could not parse redirect url: %w", err)
	}

	path, ok := strings.CutPrefix(redirectURL.Path, "/breaks/")
	if !ok {
		return "", ErrBreakNotFound
	}

	parts := strings.Split(path, "/forecasts")
	if len(parts) != 2 {
		return "", errors.New("unexpected redirect url format")
	}

	return parts[0], nil
}

// Break returns a surf break by its slug. It returns ErrBreakNotFound for non-existent surf breaks.
func (s *Scraper) Break(slug string) (Break, error) {
	path := "/breaks/" + slug

	req, err := http.NewRequest(http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return Break{}, fmt.Errorf("could not prepare request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return Break{}, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return Break{}, ErrBreakNotFound
		}
		return Break{}, fmt.Errorf("received response with %d status code", resp.StatusCode)
	}

	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if err != nil {
		return Break{}, fmt.Errorf("could not parse response body as html: %w", err)
	}

	b, err := scrapeSurfBreak(node)
	if err != nil {
		return Break{}, fmt.Errorf("could not scrape surf break: %w", err)
	}

	return b, nil
}

func scrapeSurfBreak(n *html.Node) (Break, error) {
	navNode, ok := htmlutil.FindOne(n, htmlutil.WithIDEqual("dropformcont-nav"))
	if !ok {
		return Break{}, errors.New("could not find navigation node")
	}

	countryNode, ok := htmlutil.FindOne(navNode, htmlutil.WithIDEqual("country_id"))
	if !ok {
		return Break{}, errors.New("could not find country node")
	}

	countryNameNode, ok := htmlutil.FindOne(countryNode, htmlutil.WithAttribute("selected"))
	if !ok {
		return Break{}, errors.New("could not find country name node")
	}

	countryNameTextNode := countryNameNode.FirstChild
	if countryNameTextNode == nil {
		return Break{}, errors.New("could not find country name text node")
	}

	breakNode, ok := htmlutil.FindOne(navNode, htmlutil.WithIDEqual("location_filename_part"))
	if !ok {
		return Break{}, errors.New("could not find surf break node")
	}

	breakNameNode, ok := htmlutil.FindOne(breakNode, htmlutil.WithAttribute("selected"))
	if !ok {
		return Break{}, errors.New("could not find surf break name node")
	}

	breakNameTextNode := breakNameNode.FirstChild
	if breakNameTextNode == nil {
		return Break{}, errors.New("could not find surf break name text node")
	}

	return Break{
		Name:        breakNameTextNode.Data,
		CountryName: countryNameTextNode.Data,
	}, nil
}
