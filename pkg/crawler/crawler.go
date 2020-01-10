package crawler

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	// URLs is a slice of pages we wan't to use for the demo.
	URLs = []string{
		"https://en.wikipedia.org/wiki/C-3PO",
		"https://en.wikipedia.org/wiki/R2-D2",
		"https://en.wikipedia.org/wiki/BB-8",
		"https://en.wikipedia.org/wiki/Ash_(Alien)",
		"https://en.wikipedia.org/wiki/Bishop_(Aliens)",
		"https://en.wikipedia.org/wiki/Motorola_Droid",
		"https://en.wikipedia.org/wiki/HK-47",
		"https://en.wikipedia.org/wiki/K-2SO",
		"https://en.wikipedia.org/wiki/Obi-Wan_Kenobi",
		"https://en.wikipedia.org/wiki/Luke_Skywalker",
	}
)

// New instatiates a new instance of the webcrawler
func New(elasticClient *elasticsearch.Client, redisClient *redis.Client, logger *logrus.Logger) (statusCode int) {
	// TODO:
	//	- for each of the links in URLs
	// 		- add link to the PROCESSING set in Redis
	// 		- render the page
	// 		- get the structured data
	// 		- create a document to store in elasticSearch
	// 		- send the document to Elastic
	//		- add the link to DONE or ERROR sets
	//		- remove the link from PROCESSING
	return 200
}
