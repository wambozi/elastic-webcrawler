package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Options for the elasticsearch client
type Options struct {
	Hosts                 []string
	User, Password, Index string
}

// Document type for indexing in Elasticsearch
type Document struct {
	Index, DocumentID string
	Body              *strings.Reader
}

// ElasticClient returns the elasticsearch client
func ElasticClient(o *Options) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: o.Hosts,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	log.Println("Creating Elasticsearch Client... Done.")
	return es, nil
}

// GetInfo returns elasticsearch cluster info
func GetInfo(es *elasticsearch.Client) *esapi.Response {
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting elastic health: %s", err.Error())
	}
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	return res
}

// IndexDocuments concurrently indexes documents and returns the response
func IndexDocuments(es *elasticsearch.Client, documents []Document) {
	var (
		r  map[string]interface{}
		wg sync.WaitGroup
	)
	for i, d := range documents {
		wg.Add(1)

		go func(i int, d Document) {
			defer wg.Done()

			// Define the request object
			req := esapi.IndexRequest{
				Index:      d.Index,
				DocumentID: d.DocumentID,
				Body:       d.Body,
				Refresh:    "true",
			}

			// Perform the request with the provided client
			res, err := req.Do(context.Background(), es)
			if err != nil {
				// Fatal error if the indexing request throws
				log.Fatalf("Error getting index response: %s", err)
			}
			defer res.Body.Close()

			// If the response is an error, print but proceed
			if res.IsError() {
				log.Printf("[%s] Error indexing document ID=%s", res.Status(), d.DocumentID)
			} else {
				// Deserialize the response into a map
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Printf("Error deserializing the response object: %s", err)
				} else {
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, d)
	}
}

// SendToElastic takes a document as a bytes.Reader and the indexName and sends the document to
// Elasticsearch using the client provided
func SendToElastic(data *bytes.Reader, indexName string, elasticClient *elasticsearch.Client) string {

	res, err := elasticClient.Bulk(data, elasticClient.Bulk.WithIndex(indexName))
	errStr := ElasticErrorHandler(*res, err)

	res.Body.Close()
	return errStr
}

// ElasticErrorHandler accepts the elastic API response/error and returns a string with the errors, if there are any
func ElasticErrorHandler(res esapi.Response, err error) string {
	var (
		blk    *bulkResponse
		raw    map[string]interface{}
		errStr string
	)

	if err != nil {
		errStr = fmt.Sprintf("Failure to index log message: %s", err)
		return errStr
	}

	if res.IsError() {
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			errStr = fmt.Sprintf("Failure to parse response body: %s", err)
			return errStr
		}
		errStr = fmt.Sprintf("Request Error: [%d] %s: %s",
			res.StatusCode,
			raw["error"].(map[string]interface{})["type"],
			raw["error"].(map[string]interface{})["reason"],
		)
		return errStr
	}

	if err := json.NewDecoder(res.Body).Decode(&blk); err != nil {
		errStr = fmt.Sprintf("Failure to to parse response body: %s", err)
		return errStr
	}

	log.Println(blk.Items)

	for _, d := range blk.Items {
		if d.Index.Status > 201 {
			errStr = fmt.Sprintf("Items Error: [%d]: %+v",
				d.Index.Status,
				d.Index.Error,
			)
			return errStr
		}
	}

	return errStr
}
