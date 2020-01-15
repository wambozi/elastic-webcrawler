package serving

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
	"github.com/wambozi/elastic-webcrawler/m/pkg/crawler"
)

func TestHandleIndex(t *testing.T) {
	r := httprouter.New()
	l := logrus.New()
	ac := clients.CreateAppsearchClient(ase, token, api)
	cfg := clients.GenerateElasticConfig([]string{ee}, username, password)
	ec, err := clients.CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}

	type results struct {
		Body       string
		StatusCode int
	}

	tests := map[string]struct {
		server     *Server
		statusCode int
		body       string
		log        *logrus.Logger
	}{
		"elasticsearch": {server: &Server{AppsearchClient: ac, ElasticClient: ec, Router: r, Log: l}, statusCode: 202, body: `{"status":201,"url":"https://www.example.com","type":"elasticsearch","index":"test"}`},
		// TODO: figure out why this panics...
		// "app-search":    {server: &Server{AppsearchClient: ac, ElasticClient: ec, Router: r, Log: l}, statusCode: 202, body: `{"status":201,"url":"https://www.example.com","type":"app-search","engine":"test"}`},
		// "bad-request":   {server: &Server{AppsearchClient: ac, ElasticClient: ec, Router: r, Log: l}, statusCode: 400, body: `{"status":201,"url":"https://www.example.com","type":"test","index":"test"}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.server.routes()
			b := crawler.CrawlRequest{
				Index: "test",
				URL:   "https://www.example.com",
				Type:  "elasticsearch",
			}
			bodyJSON, err := json.Marshal(b)
			encodedBody := bytes.NewReader(bodyJSON)
			req, err := http.NewRequest("POST", "/crawl", encodedBody)
			if err != nil {
				t.Fatalf("new request error: %+v", err)
			}
			w := httptest.NewRecorder()
			tc.server.Router.ServeHTTP(w, req)

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(w.Result().Body)
			if err != nil {
				t.Fatalf("could not read response body: %+v", err)

			}

			body := buf.String()

			gotRes := results{
				Body:       body,
				StatusCode: w.Result().StatusCode,
			}

			wantRes := results{
				Body:       tc.body,
				StatusCode: tc.statusCode,
			}

			diff := cmp.Diff(gotRes, wantRes)
			if diff != "" {
				t.Fatalf(diff)
			}
		})

	}

}
