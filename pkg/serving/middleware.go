package serving

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

//Middleware is transparent, and since it's just another handler function, the call to the next handler h(w,r) can be done anywhere in the midst of the middleware function's execution.

func (s *Server) execDurLog(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		elapsed := time.Since(start).Nanoseconds()
		s.Log.Infof("Response took : %d nanoseconds for %s %s", elapsed, r.Method, r.RequestURI)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	log    *logrus.Logger
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.log.Infof("Response status: %d", code)
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.log.Infof("Response body: %s", string(b))
	return r.ResponseWriter.Write(b)
}

func (s *Server) reqResLog(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Log.Infof("Request: %s %s", r.Method, r.RequestURI)
		s.Log.Infof("Request headers: %+v", r.Header)

		if r.Body != nil {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			s.Log.Infof("Request body: %s", string(bodyBytes))
			if err != nil {
				s.Log.Errorf("Could not ready request body: %w", err)
			}
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Initialize the status to 200 in case WriteHeader is not called explicitly in subsequent handlers (it defaults to 200)
		rec := responseRecorder{w, 200, s.Log}

		// Pass responseRecorder to subsequent handlers so that its implementations of Write() and WriteHeader() are used
		h(&rec, r)

	}
}
