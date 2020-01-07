package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// GetElasticCreds returns the elasticsearch credentials from AWS SecretsManager
func GetElasticCreds(secret SecretInput, env string) (creds ElasticsearchCredentials, err error) {
	var elasticCreds ElasticsearchCredentials

	// get the secret object from SecretsManager
	result, err := secret.Client.GetSecretValue(secret.Input)
	if err != nil {
		return elasticCreds, err
	}

	// declare a variable that has the Secret type that can be unmarshalled into
	sec := Secret{}

	// Determine if SecretString is a string pointer or binary to decode
	if result.SecretString != nil {
		sec.Secret = *result.SecretString

		// Unmarshall Secret string
		err = json.Unmarshal([]byte(sec.Secret), &elasticCreds)
		if err != nil {
			return elasticCreds, err
		}

		return elasticCreds, nil
	}

	return elasticCreds, nil
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

func logError(logger *logrus.Logger, err error, requestID string) {
	errFormat := "[error] in %s[%s:%d]"
	pc, fn, line, _ := runtime.Caller(1)
	logger.WithFields(logrus.Fields{
		"caller":    fmt.Sprintf(errFormat, runtime.FuncForPC(pc).Name(), fn, line),
		"requestID": requestID,
	}).Error(err.Error())
}

// All takes a reader object and returns links from `href` elements from the DOM.
// It does not close the reader.
func All(httpBody io.Reader) []string {
	links := []string{}
	col := []string{}
	page := html.NewTokenizer(httpBody)

	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					tl := trimAnchor(attr.Val)
					col = append(col, tl)
					resolv(&links, col)
				}
			}
		}
	}
}

func trimAnchor(link string) string {
	if strings.Contains(link, "#") {
		var index int
		for n, str := range link {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return link[:index]
	}
	return link
}

func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

func resolv(sl *[]string, ml []string) {
	for _, str := range ml {
		if check(*sl, str) == false {
			*sl = append(*sl, str)
		}
	}
}
