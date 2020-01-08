package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

// CreateElasticClient returns the Elasticsearch client used by the function to connect to
// Elasticsearch, using the config provided
func CreateElasticClient(cfg elasticsearch.Config) (client *elasticsearch.Client, err error) {
	client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return client, err
	}

	return client, nil
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
