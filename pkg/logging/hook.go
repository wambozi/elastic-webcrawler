package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
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

var (
	// ErrCannotCreateIndex Fired if the index is not created
	ErrCannotCreateIndex = fmt.Errorf("cannot create index")
)

// NewAsyncElasticHook creates new hook with asynchronous log.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewAsyncElasticHook(client *elasticsearch.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewAsyncElasticHookWithFunc(client, host, level, func() string { return index })
}

// NewAsyncElasticHookWithFunc creates new asynchronous hook with
// function that provides the index name.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// indexFunc - function providing the name of index
func NewAsyncElasticHookWithFunc(client *elasticsearch.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, asyncFireFunc)
}

func newHookFuncAndFireFunc(client *elasticsearch.Client, host string, level logrus.Level, indexFunc IndexNameFunc, fireFunc fireFunc) (*ElasticHook, error) {
	var levels []logrus.Level
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())

	// Check to see if the index exists, and if not create it
	exists, err := client.Indices.Exists([]string{indexFunc()})
	if err != nil {
		// Handle error
		cancel()
		return nil, err
	}
	if exists.StatusCode != 200 {
		createIndex, err := client.Indices.Create(indexFunc())
		if err != nil {
			cancel()
			return nil, err
		}
		if createIndex.StatusCode != 200 {
			cancel()
			return nil, ErrCannotCreateIndex
		}
	}

	return &ElasticHook{
		client:    client,
		host:      host,
		index:     indexFunc,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
		fireFunc:  fireFunc,
	}, nil
}

// Fire is required to implement a Logrus hook interface
// https://godoc.org/github.com/sirupsen/logrus#Hook
func (hook *ElasticHook) Fire(entry *logrus.Entry) error {
	return hook.fireFunc(entry, hook)
}

func asyncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	go syncFireFunc(entry, hook)
	return nil
}

func createDocument(entry *logrus.Entry, hook *ElasticHook) *document {
	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	return &document{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}
}

func syncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	var (
		buf    bytes.Buffer
		b      *bytes.Reader
		raw    map[string]interface{}
		errStr string
	)

	data, err := json.Marshal(*createDocument(entry, hook))
	if err != nil {
		return err
	}

	data = append(data, "\n"...)
	buf.Grow(len(data))
	buf.Write(data)

	b = bytes.NewReader(buf.Bytes())
	req := esapi.IndexRequest{
		Index:   hook.index(),
		Body:    b,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), hook.client)
	if err != nil {
		return err
	}
	if res.IsError() {
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			errStr = fmt.Sprintf("Failure to parse response body: %s", err)
			return errors.New(errStr)
		}
		errStr = fmt.Sprintf("  Error: [%d] %s: %s",
			res.StatusCode,
			raw["error"].(map[string]interface{})["type"],
			raw["error"].(map[string]interface{})["reason"],
		)
		return errors.New(errStr)
	}

	return err
}

// Levels Required for logrus hook implementation
func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}

// Cancel all calls to elastic
func (hook *ElasticHook) Cancel() {
	hook.ctxCancel()
}
