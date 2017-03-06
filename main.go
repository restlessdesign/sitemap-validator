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
	Sitemaps []Sitemap
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod"`
}

type UrlSet struct {
	Urls []SitemapURL
}

type SitemapURL struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod"`
	Changefreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("You must pass a sitemap index URL argument!")
	}

	// Get sitemap index from arguments
	sitemapIndexUrl := os.Args[1]

	// Make request to sitemap index
	resp, err := http.Get(sitemapIndexUrl)
	if err != nil || resp.StatusCode > 200 {
		log.Println("Unable to load sitemap: ", sitemapIndexUrl)
		log.Fatal(err)
	}

	parseSitemapResp(resp)
}

// Parse sitemap XML response for other sitemaps
func parseSitemapResp(resp *http.Response) {
	// Ensure that read operation cleans up after itself
	defer resp.Body.Close()

	// Load response into a byte array
	log.Println("Loading sitemap index…")
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading sitemap index")
		log.Fatal(err)
	}

	log.Println(string(respBytes))

	// Parse XML from response byte array
	v := &SitemapIndex{}
	xmlErr := xml.Unmarshal(respBytes, &v)
	if xmlErr != nil {
		log.Println("Error parsing XML response")
		log.Fatal(xmlErr)
	}

	log.Print(v)
}

// Build tree of URLs
// Issue a HEAD request for each URL, IGNORING redirects
// For each link in the sitemap, save parent sitemap path, sitemap link path, and HTTP status (200, 301, 404, 500, etc.) to CSV
