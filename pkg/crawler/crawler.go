package crawler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
)

var visited = make(map[string]bool)
var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

// Retries represents the Error retries for Crawler requests
type Retries struct {
	Enabled bool `json:"enabled,omitempty"`
	Number  int  `json:"number,omitempty"`
}

// Source represents the data scraped from the source of the HTML body on the page
type Source struct {
	h1 []string
	h2 []string
	h3 []string
	h4 []string
	p  []string
}

// Meta represents the data scraped from the metadata of the page
type Meta struct {
	OgImage  string
	Title    string
	Desc     string
	Keywords string
}

// RenderedPage represents the structred data scraped from the page
type RenderedPage struct {
	URI    string
	Source map[string][]string
	Meta   Meta
}

// CrawlRequest represents the request to the /crawl route
type CrawlRequest struct {
	Index    string `json:"index"`
	URL      string `json:"url"`
	OnDomain bool   `json:"on_domain,omitempty"`
}

// Init initializes a new crawl
func Init(elasticClient *elasticsearch.Client, cr CrawlRequest, logger *logrus.Logger) (statusCode int) {
	validURL, err := url.ParseRequestURI(cr.URL)
	if err != nil {
		return 400
	}

	domain := validURL.Hostname()

	go func(u string, d string, i string, e *elasticsearch.Client, l *logrus.Logger) { Crawl(u, d, i, e, l) }(validURL.String(), domain, cr.Index, elasticClient, logger)

	return 201
}

func appendToSlice(sl *[]string, ml string) {
	*sl = append(*sl, ml)
}

// Crawl does the crawling
func Crawl(uri string, domain string, index string, elasticClient *elasticsearch.Client, logger *logrus.Logger) {
	c := colly.NewCollector(
		colly.AllowedDomains(domain),
	)

	// Callback for when a scraped page contains an article element
	c.OnHTML("article", func(e *colly.HTMLElement) {
		page := RenderedPage{
			URI: e.Request.URL.String(),
			Meta: Meta{
				Title: e.DOM.Find("title").Text(),
			},
			Source: make(map[string][]string),
		}

		metaTags := e.DOM.ParentsUntil("~").Find("meta")
		metaTags.Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if strings.EqualFold(name, "description") {
				content, _ := s.Attr("content")
				page.Meta.Desc = content
			}
			if strings.EqualFold(name, "keywords") {
				content, _ := s.Attr("content")
				page.Meta.Keywords = content
			}
		})

		for _, el := range []string{"h1", "h2", "h3", "h4", "p"} {
			e.DOM.Find(el).Each(func(_ int, s *goquery.Selection) {
				page.Source[el] = append(page.Source[el], s.Text())
			})
		}

		doc, err := CreateDocument(index, page)
		if err != nil {
			logger.Error(err)
		}

		clients.IndexDocument(elasticClient, doc, logger)
	})

	// Callback for links on scraped pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: (1 * time.Second) / 3,
	})

	c.OnRequest(func(r *colly.Request) {
		logger.Infof("Visiting: %s", r.URL.String())
	})

	c.Visit(uri)
}

func fixURL(href, base string) (URL string, err error) {
	uri, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	uri = baseURL.ResolveReference(uri)
	return uri.String(), nil
}

// CreateDocument returns the document to be indexed in Elasticsearch or an error
func CreateDocument(i string, p RenderedPage) (doc clients.Document, err error) {
	idBytes := md5.Sum([]byte(p.URI))
	idHash := hex.EncodeToString(idBytes[:])
	bodyJSON, err := json.Marshal(p)
	if err != nil {
		return doc, err
	}
	r := bytes.NewReader(bodyJSON)
	doc = clients.Document{
		Index:      i,
		DocumentID: idHash,
		Body:       r,
	}

	return doc, nil
}

// trimHash slices a hash # from the link
func trimHash(l string) string {
	if strings.Contains(l, "#") {
		var index int
		for n, str := range l {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return l[:index]
	}
	return l
}

// check looks to see if a url exits in the slice.
func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

// validateURI takes a string, validates that it's a valid URI
func validateURI(str string) bool {
	_, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	return true
}

// checkDomain takes a URI as a string and validates that it's on a provided domain
func checkDomain(uri string, onDomain bool, domain string) bool {
	parsedURI, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}
	if onDomain {
		if parsedURI.Hostname() == domain {
			return true
		}
	}
	return true
}

// onlyWebPages checks that the link provided is a webpage and not a link to a file
func onlyWebPages(uri string) (detections []int) {
	var invalidPaths []string

	invalidPaths = []string{".png", ".jpeg", ".jpg", ".ogg", ".woff", ".pdf", ".gif", ".tiff", ".svg"}
	for _, p := range invalidPaths {
		if strings.Contains(uri, p) {
			detections = append(detections, 1)
		}
	}

	return
}

// resolv adds links to the link slice and insures that there is no repetition
// in our collection.
func resolv(sl *[]string, ml []string, onDomain bool, domain string) {
	for _, str := range ml {
		if check(*sl, str) == false && validateURI(str) == true && len(onlyWebPages(str)) == 0 {
			if checkDomain(str, onDomain, domain) == true {
				*sl = append(*sl, str)
			}
		}
	}
}
