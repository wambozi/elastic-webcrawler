package serving

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wambozi/elastic-webcrawler/m/pkg/crawler"
)

// Response is a concrete representation of the response to the client calling the crawl
type Response struct {
	Status int    `json:"status"`
	URL    string `json:"url"`
	Type   string `json:"type"`
	Index  string `json:"index,omitempty"`
	Engine string `json:"engine,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleCrawl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var b crawler.CrawlRequest
		w.Header().Set("Content-Type", "application/json")

		err := json.NewDecoder(r.Body).Decode(&b)
		if err != nil {
			s.Log.Error(err)
			er := errorResponse{Error: err.Error()}
			ers, _ := json.Marshal(er)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(ers)
			return
		}

		if b.Type != "app-search" && b.Type != "elasticsearch" {
			err := errorResponse{Error: fmt.Sprintf("Crawl type of: %s is not supported. Must be 'app-search' or 'elasticsearch'", b.Type)}
			ers, _ := json.Marshal(err)

			w.WriteHeader(http.StatusBadRequest)
			w.Write(ers)
			return
		}

		status := crawler.Init(s.ElasticClient, s.AppsearchClient, b, s.Log)

		res := Response{}

		if b.Type == "elasticsearch" {
			res = Response{Status: status, URL: b.URL, Type: "elasticsearch", Index: b.Index}
		}

		if b.Type == "app-search" {
			res = Response{Status: status, URL: b.URL, Type: "app-search", Engine: b.Engine}
		}

		response, err := json.Marshal(res)
		if err != nil {
			es := fmt.Sprintf("Failed to marshal %+v", res)
			er := errorResponse{Error: es}
			ers, _ := json.Marshal(er)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(ers)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write(response)
	}
}
