package serving

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/logger"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/crawling"
	"github.com/wambozi/elastic-webcrawler/m/pkg/validating"
)

// set the application/JSON header with a function instead of repeating this constant
// every time
func setJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func writeResponse(w http.ResponseWriter, statusCode int, body []byte) {
	setJSONHeader(w)
	w.WriteHeader(statusCode)
	w.Write(body)
}

func writeErrors(w http.ResponseWriter, l *logrus.Logger, sc int, errs []error) {
	if len(errs) == 1 {
		firstErr := errs[0]
		unwrappedErr0 := errors.Unwrap(firstErr)
		if unwrappedErr0 != nil {
			l.Error(unwrappedErr0.Error())
		}
		er := errorResponse{Error: firstErr.Error()}
		ers, _ := json.Marshal(er)
		writeResponse(w, sc, ers)
		return
	}

	er := []errorResponse{}

	for _, e := range errs {
		unwrappedErr1 := errors.Unwrap(e)
		if unwrappedErr1 != nil {
			l.Error(unwrappedErr1.Error())
		}
		er = append(er, errorResponse{Error: e.Error()})
	}
	ers, _ := json.Marshal(er)

	writeResponse(w, sc, ers)
}

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

func (s *Server) handleAppSearchCrawl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var b validating.CrawlRequest
		l := logrus.StandardLogger()
		w.Header().Set("Content-Type", "application/json")

		err := json.NewDecoder(r.Body).Decode(&b)
		if err != nil {
			logger.Error(err)
			er := errorResponse{Error: err.Error()}
			ers, _ := json.Marshal(er)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(ers)
			return
		}

		validatedRequest, errs := validating.ValidateAppSearchRequest(&b)
		if errs != nil {
			writeErrors(w, l, 400, errs)
		}

		status := crawling.InitAppSearch(s.AppsearchClient, *validatedRequest, s.Log)
		res := Response{Status: status, URL: b.URL, Type: "app-search", Engine: b.Engine}

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

func (s *Server) handleElasticCrawl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var b validating.CrawlRequest
		l := logrus.StandardLogger()
		w.Header().Set("Content-Type", "application/json")

		err := json.NewDecoder(r.Body).Decode(&b)
		if err != nil {
			logger.Error(err)
			er := errorResponse{Error: err.Error()}
			ers, _ := json.Marshal(er)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(ers)
			return
		}

		validatedRequest, errs := validating.ValidateElasticsearchRequest(&b)
		if errs != nil {
			writeErrors(w, l, 400, errs)
		}

		status := crawling.InitElastic(s.ElasticClient, *validatedRequest, s.Log)
		res := Response{Status: status, URL: b.URL, Type: "elasticsearch", Index: b.Index}

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
