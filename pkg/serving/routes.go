package serving

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/conf"
	"github.com/wambozi/elastic-webcrawler/m/pkg/connecting"
)

//Server defines storage and a router
type Server struct {
	AppsearchClient *connecting.AppsearchClient
	ElasticClient   *elasticsearch.Client
	Router          *httprouter.Router
	Log             *logrus.Logger
}

//NewServer sets up storage, router and routes
func NewServer(c *conf.Configuration, ac *connecting.AppsearchClient, ec *elasticsearch.Client, r *httprouter.Router, log *logrus.Logger) *Server {
	server := &Server{AppsearchClient: ac, ElasticClient: ec, Router: r, Log: log}
	server.routes()
	return server
}

//NewHTTPServer provides a server setup based on config values
func (s *Server) NewHTTPServer(c *conf.Configuration) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", c.Server.Port),
		Handler:           s.Router,
		ReadHeaderTimeout: time.Duration(c.Server.ReadHeaderTimeoutMillis) * time.Millisecond,
	}
}

//Begin starts an httpServer with configuration values and server values
func (s *Server) Begin(hs *http.Server, wg *sync.WaitGroup, once *sync.Once, signals chan os.Signal, errs chan<- error) {
	wg.Add(1)
	defer func(wgp *sync.WaitGroup, onceP *sync.Once, errsP chan<- error) {
		wgp.Done()
		wgp.Wait() //Wait until other goroutine(s) are done before trying to close channel
		closeChannel(onceP, errsP)
	}(wg, once, errs)

	go s.shutdownOnSignal(hs, wg, once, signals, errs)

	err := hs.ListenAndServe()
	if err != nil && err != http.ErrServerClosed { //ListenAndServe always returns non-nil error. http.ErrServerClosed is the "expected" error if shutdown/closed properly.
		errs <- fmt.Errorf("ListenAndServe error: %w", err)
	}
}

func (s *Server) shutdownOnSignal(serv *http.Server, wg *sync.WaitGroup, once *sync.Once, signals chan os.Signal, errs chan<- error) {
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	sig := <-signals
	s.Log.Infof("Received signal : %v. Server shutting down.", sig)
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func(cnc context.CancelFunc, wgp *sync.WaitGroup, onceP *sync.Once, errsP chan<- error, logP *logrus.Logger) {
		//extra cleanup can be done here (e.g. closing database connection)
		logP.Infof("Extra cleanup - closing the following connection : %+v", s.ElasticClient)

		cnc()

		wgp.Done()
		wgp.Wait() //Wait until other goroutine(s) are done before trying to close channel
		closeChannel(onceP, errsP)
		signal.Stop(signals)
		close(signals)
	}(cancel, wg, once, errs, s.Log)

	err := serv.Shutdown(ctxShutDown)
	if err != nil && err != http.ErrServerClosed { //http.ErrServerClosed is the "expected" error (returned immediately) if shutdown properly
		//Error from closing listeners or context timeout
		errs <- fmt.Errorf("Shutdown error: %w", err)
	}
}

func closeChannel(once *sync.Once, channel chan<- error) {
	once.Do(
		func() {
			close(channel)
		})
}

func (s *Server) routes() {
	s.Router.HandlerFunc("POST", "/crawl", s.execDurLog(s.reqResLog(s.handleCrawl())))
}
