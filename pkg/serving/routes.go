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

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/conf"
)

//Persister persists data
type Persister interface {
	Persist(someData string) (string, error)
}

//DatastoreNamer provides the name of the datastore
type DatastoreNamer interface {
	fmt.Stringer
}

//Storer persists data and provides information about the underlying datastore
type Storer interface {
	Persister
	DatastoreNamer
}

//Persist saves data in a datastore
func (st *Storage) Persist(someData string) (string, error) {
	return fmt.Sprintf("I persisted - %s in - %s", someData, st.SomeDatabase), nil
}

//String provides information about Storage's underlying datastore
func (st *Storage) String() string {
	return st.SomeDatabase
}

//Storage defines datastore
type Storage struct {
	SomeDatabase string
}

//Server defines storage and a router
type Server struct {
	Storage Storer
	Router  *httprouter.Router
	Log     *logrus.Entry
}

//NewServer sets up storage, router and routes
func NewServer(c *conf.Configuration, s *Storage, r *httprouter.Router, log *logrus.Entry) *Server {
	server := &Server{Storage: s, Router: r, Log: log}
	server.routes()
	return server
}

//NewStorage provides a connection to the application's underlying data store
func NewStorage(c *conf.StorageConfiguration) (*Storage, error) {
	//here you would use storage configuration info to create a postgres connection, prepare statements (if needed), etc, and then return *sql.DB.
	//Returning *storage for illustration purposes only.
	return &Storage{
		SomeDatabase: fmt.Sprintf("Successful connection to : %s", c.ConnectionURI),
	}, nil
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

	defer func(cnc context.CancelFunc, wgp *sync.WaitGroup, onceP *sync.Once, errsP chan<- error, logP *logrus.Entry) {
		//extra cleanup can be done here (e.g. closing database connection)
		logP.Infof("Extra cleanup - closing the following connection : %s", s.Storage.String())

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
	s.Router.HandlerFunc("GET", "/", s.execDurLog(s.corrIDHeaderLog(s.reqResLog())))
	s.Router.HandlerFunc("POST", "/split", s.execDurLog(s.corrIDHeaderLog(s.reqResLog())))

	s.Router.ServeFiles("/docs/*filepath", http.Dir("./pkg/serving/swagger/"))
}
