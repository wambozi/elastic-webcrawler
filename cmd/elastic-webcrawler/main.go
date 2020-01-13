package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/conf"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
	"github.com/wambozi/elastic-webcrawler/m/pkg/logging"
	"github.com/wambozi/elastic-webcrawler/m/pkg/serving"
)

var (
	err error
)

// entrypoint
func main() {
	logger := logrus.New()

	err := run(logger)
	if err != nil {
		logger.Errorf("stdErr: %+v , error: %v", os.Stderr, err)
		os.Exit(1)
	}
}

// Run executes the lambda function
func run(logger *logrus.Logger) error {
	logger.Info("No .env file found. Using viper to get config values.")
	e := conf.GetEnvironment()
	c, err := conf.Setup(e)
	if err != nil {
		return err
	}

	elasticConfig := clients.GenerateElasticConfig([]string{c.Elasticsearch.Endpoint}, c.Elasticsearch.Username, c.Elasticsearch.Password)
	elasticClient, err := clients.CreateElasticClient(elasticConfig)
	if err != nil {
		return err
	}

	ipAddr, err := logging.GetIPAddr()
	if err != nil {
		return err
	}

	// Create async elasticsearch hook for logrus
	hook, err := logging.NewAsyncElasticHook(elasticClient, ipAddr.String(), logrus.DebugLevel, "elastic-webcrawler")
	if err != nil {
		return err
	}
	logger.Hooks.Add(hook)
	logger.Info("Initialized")

	// now that we've added the ELastic hook to the logger, we'll log errs as they occur so they show
	// up in Elasticsearch but still return them so they are logged in the console

	logger.Infof("Configuration : %+v", c)

	r := httprouter.New()

	server := serving.NewServer(c, elasticClient, r, logger)
	logger.Infof("Server components: %+v", server)

	httpServer := server.NewHTTPServer(c)
	logger.Infof("httpServer : %+v", httpServer)

	var doOnce sync.Once               //for closing the error channel
	var wg sync.WaitGroup              //for ensuring graceful shutdown
	signals := make(chan os.Signal)    //for shutdown signals
	httpSvrErrs := make(chan error, 2) //for http server errors

	wg.Add(1)
	go server.Begin(httpServer, &wg, &doOnce, signals, httpSvrErrs)

	wg.Wait()
	logger.Infof("Server stopped")

	if len(httpSvrErrs) > 0 {
		var errs []string

		for v := range httpSvrErrs {
			errs = append(errs, v.Error())
		}

		logger.Error(strings.Join(errs, "  |  "))
		return fmt.Errorf(strings.Join(errs, "  |  "))
	}

	return nil
}
