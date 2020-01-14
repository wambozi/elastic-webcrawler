package clients

import (
	"net"
	"net/http"
	"time"
)

// AppsearchDocument represents the document sent to App Search
type AppsearchDocument struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	URI         string              `json:"uri"`
	Source      map[string][]string `json:"source"`
	OgImage     string              `json:"ogimage"`
	Title       string              `json:"title"`
	Keywords    string              `json:"keywords"`
}

// AppsearchClient represents the HTTP client and configs used to send requests to App Search
type AppsearchClient struct {
	Client   *http.Client
	Token    string
	Endpoint string
	API      string
}

// CreateAppsearchClient creates the client for App Search
func CreateAppsearchClient(e string, t string, a string) *AppsearchClient {
	client := &http.Client{
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
	return &AppsearchClient{
		Client:   client,
		Token:    t,
		Endpoint: e,
		API:      a,
	}
}
