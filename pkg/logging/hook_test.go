package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/connecting"
)

var (
	red   = color.FgRed.Render
	green = color.FgGreen.Render
)

type NewHookFunc func(client *elasticsearch.Client, host string, level logrus.Level, index string) (*ElasticHook, error)

type DocumentCount struct {
	Count int `json:"count"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func TestAsyncHook(t *testing.T) {
	hookTest(NewAsyncElasticHook, "async-log", t)
}

func hookTest(hookfunc NewHookFunc, indexName string, t *testing.T) {
	endpoint := "http://" + os.Getenv("ELASTICSEARCH_IP") + ":9200"
	elasticConfig := elasticsearch.Config{
		Addresses: []string{endpoint},
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

	// Test that elastic is reachable before coninuting with the test
	if _, err := http.Get(endpoint); err != nil {
		t.Fatal("Elastic not reachable")
	}

	// Create the elasticsearch client using the config defined above
	elasticClient, err := connecting.CreateElasticClient(elasticConfig)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	// Delete the index and ignore the response
	_, err = elasticClient.Indices.Delete([]string{indexName})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	// Create logrus hook using the elastic client
	hook, err := hookfunc(elasticClient, "localhost", logrus.DebugLevel, indexName)
	if err != nil {
		log.Panic(err)
	}
	logrus.AddHook(hook)

	// Create 100 sample log messages and log them
	samples := 100
	for index := 0; index < samples; index++ {
		logrus.Infof("Testing msg %d", time.Now().Unix())
	}

	// Allow time for data to be processed.
	time.Sleep(10 * time.Second)

	// Declare an object to Unmarshal the response object into
	res := new(DocumentCount)

	// Get the document count from the index
	e := getJSON(endpoint+"/"+indexName+"/_count?format=json", res)

	// Handle errors
	if e != nil {
		fmt.Print(e.Error())
		t.Fatalf("Not able to get document count for indexName: %s", indexName)
	}

	// Document count should equal the number of samples we created above
	if res.Count != samples {
		t.Errorf("\n%s:\n\n%d\n\n%s:\n\n%d", green("[expected]"), samples, red("[actual]"), res.Count)
	}

	hook.Cancel()
}
