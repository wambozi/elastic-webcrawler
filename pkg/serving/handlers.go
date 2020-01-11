package serving

import (
	"net/http"

	"github.com/wambozi/elastic-webcrawler/m/pkg/crawler"
)

func (s *Server) handleCrawl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		crawler.New(s.EventEmitter, s.ElasticClient, s.RedisClient, s.Log)
		w.WriteHeader(http.StatusOK)
	}
}
