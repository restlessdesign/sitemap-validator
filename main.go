// This is a utility script for building a table of URLs and their associated
// response codes within a given sitemap. If the sitemap index contains links to
// other sitemaps, it will traverse through until it comes across actual sitemap
// URLs to make HEAD requests to and determine their HTTP response codes.
//
// Usage:
// $ go run main.go https://www.compass.com/sitemaps/index/
package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	XMLNS   string   `xml:"xmlns,attr"`

	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"sitemap"`

	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type SitemapURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`

	URLs []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	XMLName xml.Name `xml:"url"`

	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

type SitemapValidationResponse struct {
	Name    string
	Results []URLValidationResponse
}

type URLValidationResponse struct {
	URL        string
	StatusCode int
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("You must pass a sitemap index URL argument!")
	}

	loadSitemapIndex(os.Args[1])
}

func loadSitemapIndex(url string) {
	resp, err := http.Get(url)

	log.Printf("(%v) %v \n", resp.Status, url)

	if err != nil || resp.StatusCode > 200 {
		log.Printf("Unable to load sitemap index: %v\n", url)
		log.Fatal(err)
	}

	parseSitemapIndex(resp)
}

func parseSitemapIndex(resp *http.Response) {
	// Ensure that read operation cleans up after itself
	defer resp.Body.Close()

	// Load response into a byte array
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading sitemap index")
		log.Fatal(err)
	}

	// Parse XML from response byte array
	v := &SitemapIndex{}
	xmlErr := xml.Unmarshal(respBytes, &v)
	if xmlErr != nil {
		log.Println("Error parsing XML response")
		log.Fatal(xmlErr)
	}

	ch := make(chan SitemapValidationResponse, len(v.Sitemaps))
	for _, sitemap := range v.Sitemaps {
		go loadSitemap(sitemap.Loc, ch)
	}

	// Iterate over channel output
	results := make([]SitemapValidationResponse, len(v.Sitemaps))
	for i, _ := range v.Sitemaps {
		results[i] = <-ch
	}

	log.Println("All sitemaps loaded")
	log.Print(results)
}

func loadSitemap(url string, results chan SitemapValidationResponse) {
	resp, err := http.Get(url)

	log.Printf("(%v) %v \n", resp.Status, url)

	if err != nil || resp.StatusCode > 200 {
		log.Fatal(err)
	}

	// ch := make(chan []ValidationResponse)
	// parseSitemap(resp)

	// Notify channel of success
	results <- SitemapValidationResponse{
		Name: url,
		Results: []URLValidationResponse
	}
}

func parseSitemap(resp *http.Response) {
	// Ensure that read operation cleans up after itself
	defer resp.Body.Close()

	// Load response into a byte array
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading sitemap")
		log.Fatal(err)
	}

	// Build tree of URLs by parsing XML from response byte array
	v := &SitemapURLSet{}
	xmlErr := xml.Unmarshal(respBytes, &v)
	if xmlErr != nil {
		log.Println("Error parsing XML response")
		log.Fatal(xmlErr)
	}

	for _, url := range v.URLs {
		loadSitemapUrl(url.Loc)
	}
}

// func loadSitemapUrl(url string) {
// Issue a HEAD request for each URL, IGNORING redirects, spacing the requests out if at all possible
// resp, err := http.Head(url)
// }

// For each link in the sitemap, save parent sitemap path, sitemap link path, and HTTP status (200, 301, 404, 500, etc.) to CSV
