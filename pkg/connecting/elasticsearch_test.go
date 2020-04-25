package connecting

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

var (
	endpoint = "http://localhost:9200"
	username = "elastic"
	password = "changeme"
)

func TestGenerateElasticConfig(t *testing.T) {
	cfg := GenerateElasticConfig([]string{endpoint}, username, password)

	if fmt.Sprintf("%T", cfg) != "elasticsearch.Config" {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), "elasticsearch.Config", red("[actual]"), fmt.Sprintf("%T", cfg))
	}

	if cfg.Addresses[0] != endpoint {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), endpoint, red("[actual]"), cfg.Addresses[0])
	}

	if cfg.Username != username {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), username, red("[actual]"), cfg.Username)
	}

	if cfg.Password != password {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), password, red("[actual]"), cfg.Password)
	}
}

func TestCreateElasticsearchClient(t *testing.T) {
	cfg := GenerateElasticConfig([]string{endpoint}, username, password)
	client, err := CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}

	if fmt.Sprintf("%T", client) != "*elasticsearch.Client" {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), "*elasticsearch.Client", red("[actual]"), fmt.Sprintf("%T", client))
	}
}

type TestBody struct {
	Text string
}

func TestIndexDocument(t *testing.T) {
	cfg := GenerateElasticConfig([]string{endpoint}, username, password)
	client, err := CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}

	body := TestBody{
		Text: "test",
	}

	idBytes := md5.Sum([]byte("https://www.example.com"))
	idHash := hex.EncodeToString(idBytes[:])
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Unexpected error creating document for indexing: %s", err)
	}
	r := bytes.NewReader(bodyJSON)
	doc := ElasticDocument{
		Index:      "test",
		DocumentID: idHash,
		Body:       r,
	}

	_, errSlice := IndexDocument(client, doc)

	if len(errSlice) > 0 {
		t.Errorf("Unexpected error indexing documents: %v", errSlice)
	}
}
