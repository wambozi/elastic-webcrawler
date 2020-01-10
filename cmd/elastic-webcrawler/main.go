package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/conf"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
	"github.com/wambozi/elastic-webcrawler/m/pkg/logging"
	"github.com/wambozi/elastic-webcrawler/m/pkg/serving"
)

var (
	err          error
	elasticCreds clients.Credentials
	redisCreds   clients.Credentials
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
	e := conf.GetEnvironment()
	c, err := conf.Setup(e)
	if err != nil {
		return err
	}

	awsConfig := clients.AwsConfig{
		Main: aws.Config{
			Region: aws.String(c.AWS.Region),
		},
		Secrets: []clients.Secrets{
			{
				Type:   "elasticsearch",
				Secret: c.Elasticsearch.SecretName,
			},
			{
				Type:   "redis",
				Secret: c.Redis.SecretName,
			},
		},
	}

	// Use the SecretInput and AWS environment to get the credentials used to connect to Elasticsearch
	credentials, err := clients.SecretsManagerClient(awsConfig, logger)
	if err != nil {
		return err
	}

	for _, credsWrapper := range credentials {
		if credsWrapper.Type == "elasticsearch" {
			elasticCreds = credsWrapper.Credentials
		}
		if credsWrapper.Type == "redis" {
			redisCreds = credsWrapper.Credentials
		}
	}

	elasticConfig := clients.GenerateElasticConfig([]string{elasticCreds.Endpoint}, elasticCreds.Username, elasticCreds.Password)
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
	redisDB, err := strconv.Atoi(redisCreds.Database)
	if err != nil {
		logger.Error(err)
		return err
	}

	redisConfig := &clients.RedisOptions{
		Host:     redisCreds.Endpoint,
		Password: redisCreds.Password,
		Database: redisDB,
	}

	redisClient, err := clients.CreateRedisClient(redisConfig)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof("Configuration : %+v", c)

	r := httprouter.New()

	server := serving.NewServer(c, elasticClient, redisClient, r, logger)
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
