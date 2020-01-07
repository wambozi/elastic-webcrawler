package main

import (
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

func crawler(elasticClient *elasticsearch.Client, events []Event, requestID string, indexName string, logger *logrus.Logger) {
	var (
		wg sync.WaitGroup
	)

	for _, e := range events {
		wg.Add(1)

		go func(event Event) {
			defer wg.Done()

		}(e)
		wg.Wait()
	}
}
