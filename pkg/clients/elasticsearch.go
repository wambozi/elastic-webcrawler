package clients

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// ElasticDocument represents the doc that gets indexed in Elasticsearch
type ElasticDocument struct {
	Index, DocumentID string
	Body              io.Reader
}

// GenerateElasticConfig returns the elasticsearch config given the endpoint(s), username and password
func GenerateElasticConfig(endpoint []string, username string, password string) elasticsearch.Config {
	return elasticsearch.Config{
		Addresses: endpoint,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// CreateElasticClient returns the Elasticsearch client used by the function to connect to
// Elasticsearch, using the config provided
func CreateElasticClient(cfg elasticsearch.Config) (client *elasticsearch.Client, err error) {
	client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return client, err
	}

	return client, nil
}

// IndexDocument takes a document and indexes it in Elasticsearch
func IndexDocument(elasticClient *elasticsearch.Client, d ElasticDocument, logger *logrus.Logger) {
	var (
		r  map[string]interface{}
		wg sync.WaitGroup
	)
	wg.Add(1)

	go func(d ElasticDocument) {
		defer wg.Done()

		// Define the request object
		req := esapi.IndexRequest{
			Index:      d.Index,
			DocumentID: d.DocumentID,
			Body:       d.Body,
			Refresh:    "true",
		}

		// Perform the request with the provided client
		res, err := req.Do(context.Background(), elasticClient)
		if err != nil {
			// Fatal error if the indexing request throws
			logger.Errorf("Error getting index response: %s", err)
		}
		defer res.Body.Close()

		// If the response is an error, print but proceed
		if res.IsError() {
			logger.Errorf("[%s] Error indexing document ID=%s, err=%s", res.Status(), d.DocumentID, res.String())
		} else {
			// Deserialize the response into a map
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				logger.Errorf("Error deserializing the response object: %s", err)
			} else {
				logger.Infof("[%s] %s; version=%d; id=%s", res.Status(), r["result"], int(r["_version"].(float64)), d.DocumentID)
			}
		}
	}(d)
}
