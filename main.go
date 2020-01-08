package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	indexName string = "demo"
)

// entrypoint
func main() {
	lambda.Start(Handler)
}

// Run executes the lambda function
func Run(lambdaContext *lambdacontext.LambdaContext, event events.CloudwatchLogsEvent, envConfig EnvConfig, awsConfig aws.Config, logger *logrus.Logger) {
	var (
		// events        []Event
		// indexName     string
		consoleFormat string = "RequestID: %s, Error: %+v"
		elasticCreds  Credentials
		redisCreds    Credentials
	)

	// Create a new AWS Session
	sess := session.Must(session.NewSession(&awsConfig))

	// Get the AWS credentials from the environment or Shared Credentials file where the function is running
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		logger.Error(fmt.Sprintf(consoleFormat, string(lambdaContext.AwsRequestID), err))
	}

	// Create the request object for SecretsManager using the secretName from the environment
	es := SecretInput{
		Client: secretsmanager.New(sess),
		Input: &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(envConfig.elasticsearchSecret),
			VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
		},
	}

	rs := SecretInput{
		Client: secretsmanager.New(sess),
		Input: &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(envConfig.redisSecret),
			VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
		},
	}

	// Use the SecretInput and AWS environment to get the credentials used to connect to Elasticsearch
	elasticCreds, err = GetCredentials(es, envConfig.env, elasticCreds)
	if err != nil {
		logger.Error(fmt.Sprintf(consoleFormat, string(lambdaContext.AwsRequestID), err))
	}

	redisCreds, err = GetCredentials(rs, envConfig.env, redisCreds)
	if err != nil {
		logger.Error(fmt.Sprintf(consoleFormat, string(lambdaContext.AwsRequestID), err))
	}

	elasticConfig := GenerateElasticConfig([]string{elasticCreds.Endpoint}, elasticCreds.Username, elasticCreds.Password)
	redisConfig := &RedisOptions{
		Host:     redisCreds.Endpoint,
		Password: redisCreds.Username,
		Database: 0,
	}

	elasticClient, err := CreateElasticClient(elasticConfig)
	if err != nil {
		logger.Error(fmt.Sprintf(consoleFormat, string(lambdaContext.AwsRequestID), err))
	}

	redisClient, err := CreateRedisClient(redisConfig)

	events := Events(logger)
	crawler(elasticClient, redisClient, events, string(lambdaContext.AwsRequestID), indexName, logger)
}

// Handler handles the lambda event context, and runs the exec function
func Handler(ctx context.Context, event events.CloudwatchLogsEvent) {
	logger := logrus.New()

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

	// Set default region for the AWS config by getting it from environment
	// if a .env file exists, this value comes from there
	// otherwise the value is obtained from the environment, which Serverless creates from
	// the config/env.yml file in this repo
	awsConfig := aws.Config{
		Region: aws.String(config.region),
	}

	// Get the AWS Request ID from the context and log it + make it available to pass around
	lambdaContext, _ := lambdacontext.FromContext(ctx)

	// Execute the lambda function
	Run(lambdaContext, event, config, awsConfig, logger)
}
