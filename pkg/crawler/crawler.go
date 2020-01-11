package crawler

import (
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

// EventPayload represents the payload sent in the event emitted by redis actions
type EventPayload struct {
	EventEmitter  events.EventEmmiter
	RedisClient   *redis.Client
	ElasticClient *elasticsearch.Client
	Logger        *logrus.Logger
}

// Init initializes a new crawl
func Init(eventEmitter events.EventEmmiter, elasticClient *elasticsearch.Client, redisClient *redis.Client, cr CrawlRequest, logger *logrus.Logger) (statusCode int) {
	// Clear the READY set if it exists
	del := redisClient.Del(cr.URL + "_READY")
	_, err := del.Result()
	if err != nil {
		logger.Error(err)
		return 400
	}

	// add URL to READY set
	res := redisClient.SAdd(cr.URL+"_READY", cr.URL)
	addStatus, err := res.Result()
	if err != nil {
		logger.Error(err)
		return 400
	}
	if addStatus <= 0 {
		logger.Errorf("Redis failed to add link to READY set. add status: %d", addStatus)
		return 400
	}
	payload := EventPayload{
		EventEmitter:  eventEmitter,
		ElasticClient: elasticClient,
		RedisClient:   redisClient,
		Logger:        logger,
	}
	eventEmitter.Emit("READY", payload)
	return 201
}

// New instatiates a new instance of the webcrawler
func New(eventPayload EventPayload) {
	// TODO:
	// 		- render the page
	// 		- get the structured data
	// 		- create a document to store in elasticSearch
	// 		- send the document to Elastic
	//		- add the link to DONE or ERROR sets
	//		- remove the link from PROCESSING
}
