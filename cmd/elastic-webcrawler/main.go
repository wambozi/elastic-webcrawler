package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/wambozi/elastic-webcrawler/m/pkg/clients"
	"github.com/wambozi/elastic-webcrawler/m/pkg/crawler"
	"github.com/wambozi/elastic-webcrawler/m/pkg/logging"
)

var (
	err          error
	elasticCreds clients.Credentials
	redisCreds   clients.Credentials
)

// EnvConfig represents all of the environment specific information needed for the application to do its job
type EnvConfig struct {
	region              string
	env                 string
	elasticsearchSecret string
	redisSecret         string
	profile             *string
}

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

	// Load .env if its present (used for local dev)
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found.")
	}

	config := EnvConfig{
		region:              os.Getenv("REGION"),
		env:                 os.Getenv("ENVIRONMENT"),
		elasticsearchSecret: os.Getenv("ELASTICSEARCH_SECRET"),
		redisSecret:         os.Getenv("REDIS_SECRET"),
	}

	awsConfig := clients.AwsConfig{
		Main: aws.Config{
			Region: aws.String(config.region),
		},
		Secrets: []clients.Secrets{
			{
				Type:   "elasticsearch",
				Secret: config.elasticsearchSecret,
			},
			{
				Type:   "redis",
				Secret: config.redisSecret,
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

	status := crawler.New(elasticClient, redisClient, logger)

	fmt.Println(status)

	return nil
}
