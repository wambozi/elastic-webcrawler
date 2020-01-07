package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

// IndexNameFunc returns the index name
type IndexNameFunc func() string

type fireFunc func(entry *logrus.Entry, hook *ElasticHook) error

// ElasticHook represents an Elasticsearch hook for Logrus
type ElasticHook struct {
	client    *elasticsearch.Client
	host      string
	index     IndexNameFunc
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
	fireFunc  fireFunc
}

// Represents the document that gets indexed in Elasticsearch
type document struct {
	Host      string
	Timestamp string `json:"@timestamp"`
	Message   string
	Data      logrus.Fields
	Level     string
}

// SecretInput matches the secretsmanager.GetSecretValueInput struct but alows us to mock the service
type SecretInput struct {
	Client secretsmanageriface.SecretsManagerAPI
	Input  *secretsmanager.GetSecretValueInput
}

// Secret is a concrete representation of the SecretsManager response
type Secret struct {
	ARN           string     `json:"ARN"`
	CreatedDate   *time.Time `json:"CreatedDate"`
	Name          string     `json:"Name"`
	Secret        string     `json:"SecretString"`
	VersionID     string     `json:"VersionId"`
	VersionStages []string   `json:"VersionStages"`
}

// ElasticsearchCredentials represent the unmarshalled JSON object from SecretsManager
type ElasticsearchCredentials struct {
	Username string `json:"elasticsearch_username"`
	Password string `json:"elasticsearch_password"`
	Endpoint string `json:"elasticsearch_endpoint"`
}

// EnvConfig represents all of the environment specific information needed for the application to do its job
type EnvConfig struct {
	region     string
	env        string
	secretName string
	profile    *string
}

// Event defines the CloudWatch event
type Event struct {
	Timestamp time.Time
	DocID     string
	Fields    EventFields
}

// EventFields represent the fields from the event to be indexed
type EventFields struct {
	Message            LogObject
	ID                 string
	LogStream          string
	LogGroup           string
	MessageType        string
	SubscriptionFilter []string
}

// MetaObject represents what information _may_ come in from a log request
// based on what ExpressJS logs in its request logs
type MetaObject struct {
	Event map[string]interface{} `json:"apiGwEvent,omitempty"`
}

// LogObject represents what a logObject _may_ look like in a CloudWatch Event
type LogObject struct {
	Raw  map[string]interface{}
	Meta MetaObject
}

// DocumentID represents the Document ID in the Bulk request
type DocumentID struct {
	_id string
}

// DocumentMetadata represents the Document Metadata sent in the Bulk request
type DocumentMetadata struct {
	index DocumentID
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by,omitempty"`
			} `json:"error,omitempty"`
		} `json:"index"`
	} `json:"items"`
}

type sOut struct {
	file   *os.File
	writer *os.File
	reader *os.File
}
