package clients

import (
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

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
