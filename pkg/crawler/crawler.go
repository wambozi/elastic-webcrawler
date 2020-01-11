package crawler

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis"
	"github.com/kataras/go-events"
	"github.com/sirupsen/logrus"
)

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

// Init initializes a new crawl
func Init(eventEmitter events.EventEmmiter, redisClient *redis.Client, cr CrawlRequest, logger *logrus.Logger) (statusCode int) {
	fmt.Printf("%+v", cr)
	return 201
}

// New instatiates a new instance of the webcrawler
func New(eventEmitter events.EventEmmiter, elasticClient *elasticsearch.Client, redisClient *redis.Client, logger *logrus.Logger) {
	// TODO:
	//	- for each of the links in URLs
	// 		- add link to the PROCESSING set in Redis
	// 		- render the page
	// 		- get the structured data
	// 		- create a document to store in elasticSearch
	// 		- send the document to Elastic
	//		- add the link to DONE or ERROR sets
	//		- remove the link from PROCESSING
}
