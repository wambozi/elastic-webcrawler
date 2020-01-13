package crawler

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
	"golang.org/x/net/html"
)

var visited = make(map[string]bool)
var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

// CrawlRequest represents the request to the /crawl route
type CrawlRequest struct {
	Index   string  `json:"index"`
	URL     string  `json:"url"`
	Retries Retries `json:"retries,omitempty"`
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
	ogImage string
	title   string
	desc    string
}

// RenderedPage represents the structred data scraped from the page
type RenderedPage struct {
	URI    string
	Links  []string
	Source Source
	Meta   Meta
}

// Init initializes a new crawl
func Init(elasticClient *elasticsearch.Client, cr CrawlRequest, logger *logrus.Logger) (statusCode int) {
	queue := make(chan string)

	go func() { queue <- cr.URL }()

	go func() {
		for uri := range queue {
			enqueue(uri, cr.Index, queue, elasticClient, logger)
		}
	}()

	return 201
}

func enqueue(uri string, index string, queue chan string, elasticClient *elasticsearch.Client, logger *logrus.Logger) {
	logger.Infof("Fetching: %s", uri)
	visited[uri] = true
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	page := RenderPage(uri, resp.Body)
	doc, err := CreateDocument(index, page)
	if err != nil {
		logger.Error(err)
	}

	clients.IndexDocument(elasticClient, doc, logger)

	for _, link := range page.Links {
		absolute, err := fixURL(link, uri)
		if err != nil {
			logger.Error(err)
		}
		if !visited[absolute] {
			go func() { queue <- absolute }()
		}
	}
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

// RenderPage takes an io.Reader and a returns
func RenderPage(uri string, httpBody io.Reader) RenderedPage {
	var renderedPage RenderedPage
	page := html.NewTokenizer(httpBody)
	h1 := []string{}
	h2 := []string{}
	h3 := []string{}
	h4 := []string{}
	p := []string{}
	links := []string{}
	col := []string{}
	for {
		tokenTag := page.Next()
		if tokenTag == html.ErrorToken {
			renderedPage.URI = uri
			renderedPage.Links = links
			renderedPage.Source.h2 = h2
			renderedPage.Source.h3 = h3
			renderedPage.Source.h4 = h4
			renderedPage.Source.p = p
			return renderedPage
		}
		// get links
		if tokenTag == html.StartTagToken {
			token := page.Token()
			if token.DataAtom.String() == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						tl := trimHash(attr.Val)
						col = append(col, tl)
						resolv(&links, col)
					}
				}
			}
			//if the name of the element is "title"
			if token.Data == "title" {
				//the next token should be the page title
				tokenTag = page.Next()
				//just make sure it's actually a text token
				if tokenTag == html.TextToken {
					renderedPage.Meta.title = page.Token().Data
				}
			}
			//if the name of the element is "title"
			if token.Data == "h1" {
				//the next token should be the page title
				tokenTag = page.Next()
				//just make sure it's actually a text token
				if tokenTag == html.TextToken {
					h1 = append(h1, page.Token().Data)
				}
			}
			//if the name of the element is "title"
			if token.Data == "h2" {
				//the next token should be the page title
				tokenTag = page.Next()
				//just make sure it's actually a text token
				if tokenTag == html.TextToken {
					h2 = append(h2, page.Token().Data)
				}
			}
			//if the name of the element is "title"
			if token.Data == "h3" {
				//the next token should be the page title
				tokenTag = page.Next()
				//just make sure it's actually a text token
				if tokenTag == html.TextToken {
					h3 = append(h3, page.Token().Data)
				}
			}
			//if the name of the element is "title"
			if token.Data == "p" {
				//the next token should be the page title
				tokenTag = page.Next()
				//just make sure it's actually a text token
				if tokenTag == html.TextToken {
					p = append(p, page.Token().Data)
				}
			}
		}
	}
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

// resolv adds links to the link slice and insures that there is no repetition
// in our collection.
func resolv(sl *[]string, ml []string) {
	for _, str := range ml {
		if check(*sl, str) == false {
			*sl = append(*sl, str)
		}
	}
}
