package serving

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/conf"
	"github.com/wambozi/elastic-webcrawler/m/pkg/connecting"
)

var (
	ee       = "http://localhost:9200"
	username = "elastic"
	password = "changeme"
	ase      = "http://localhost:3002"
	api      = "/test/api/v1/"
	token    = "private-pq75aoza89cab89ozgd46aib"
)

func TestCloseChannel(t *testing.T) {
	var doOnce sync.Once
	ch := make(chan error)

	se := "someError"

	go func() {
		ch <- fmt.Errorf(se)
	}()
	v, ok := <-ch
	if !ok {
		t.Fatal("channel should be open")
	}

	if v.Error() != (se) {
		t.Fatalf("error should be %s", se)
	}

	closeChannel(&doOnce, ch)
	closeChannel(&doOnce, ch) //no panic, so ok

	v, ok = <-ch
	if ok {
		t.Fatal("channel should be closed")
	}

	if v != nil {
		t.Fatal("error should be nil")
	}

}

func TestNewServer(t *testing.T) {
	c := conf.Configuration{Server: conf.ServerConfiguration{Port: 8080, ReadHeaderTimeoutMillis: 3000}}
	r := httprouter.New()
	l := logrus.New()
	ac := connecting.CreateAppsearchClient(ase, token, api)
	cfg := connecting.GenerateElasticConfig([]string{ee}, username, password)
	ec, err := connecting.CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}
	actual := NewServer(&c, ac, ec, r, l)

	if actual.Router == nil {
		t.Fatalf("router should not be nil")
	}
}

func TestNewHttpServer(t *testing.T) {
	readHeaderTimeout := 3000
	port := 8080
	c := conf.Configuration{Server: conf.ServerConfiguration{Port: 8080, ReadHeaderTimeoutMillis: 3000}}
	r := httprouter.New()
	l := logrus.New()
	ac := connecting.CreateAppsearchClient(ase, token, api)
	cfg := connecting.GenerateElasticConfig([]string{ee}, username, password)
	ec, err := connecting.CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}
	s := NewServer(&c, ac, ec, r, l)

	httpServer := s.NewHTTPServer(&c)
	if httpServer == nil {
		t.Fatal("httpServer should not be nil")
	}

	rht := fmt.Sprintf("%dms", readHeaderTimeout)

	expected, err := time.ParseDuration(rht)
	if err != nil {
		t.Fatalf("could not parse readHeaderTimeout duration: %s", expected)
	}

	if httpServer.ReadHeaderTimeout != expected {
		t.Fatalf("readtimeout - expected : %+v, received : %+v", expected, httpServer.ReadHeaderTimeout)
	}

	address := fmt.Sprintf(":%d", port)

	if httpServer.Addr != address {
		t.Fatalf("httpServer.Addr - expected : %s, received : %s", address, httpServer.Addr)
	}

}

func TestShutdownOnSignal(t *testing.T) {
	var doOnce sync.Once
	var wg sync.WaitGroup
	httpSvrErrs := make(chan error, 2)
	signals := make(chan os.Signal)

	c := conf.Configuration{Server: conf.ServerConfiguration{Port: 8080, ReadHeaderTimeoutMillis: 3000}}
	r := httprouter.New()
	l := logrus.New()
	ac := connecting.CreateAppsearchClient(ase, token, api)
	cfg := connecting.GenerateElasticConfig([]string{ee}, username, password)
	ec, err := connecting.CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}
	server := NewServer(&c, ac, ec, r, l)

	httpServer := server.NewHTTPServer(&c)

	wg.Add(1)
	go server.shutdownOnSignal(httpServer, &wg, &doOnce, signals, httpSvrErrs)
	signals <- syscall.SIGINT
	wg.Wait()

	v, ok := <-httpSvrErrs
	if ok {
		t.Fatal("httpSvrErrs channel should be closed")
	}

	if v != nil {
		t.Fatal("error chan read value should be nil")
	}

	s, oks := <-signals
	if oks {
		t.Fatal("signals channel should be closed")
	}

	if s != nil {
		t.Fatal("signals chan read value should be nil")
	}
}

func TestBegin(t *testing.T) {
	var doOnce sync.Once
	var wg sync.WaitGroup
	httpSvrErrs := make(chan error, 2)
	signals := make(chan os.Signal)

	c := conf.Configuration{Server: conf.ServerConfiguration{Port: 8080, ReadHeaderTimeoutMillis: 3000}}
	r := httprouter.New()
	l := logrus.New()
	ac := connecting.CreateAppsearchClient(ase, token, api)
	cfg := connecting.GenerateElasticConfig([]string{ee}, username, password)
	ec, err := connecting.CreateElasticClient(cfg)
	if err != nil {
		t.Errorf("Unexpected error creating Elasticsearch client: %s", err)
	}
	server := NewServer(&c, ac, ec, r, l)
	httpServer := server.NewHTTPServer(&c)

	wg.Add(1)
	go server.Begin(httpServer, &wg, &doOnce, signals, httpSvrErrs)
	signals <- syscall.SIGINT
	wg.Wait()

	// TODO: test that err channel is closed
}
