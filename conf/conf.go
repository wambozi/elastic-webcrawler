package conf

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis"
	"github.com/kataras/go-events"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration holds configuration values for the application
type Configuration struct {
	Server        ServerConfiguration
	Elasticsearch ElasticOptions
	Redis         RedisOptions
}

// RedisOptions for the Redis Client
type RedisOptions struct {
	Host     string
	Port     int
	Password string
	Database int
}

// EventPayload represents the payload sent in the event emitted by redis actions
type EventPayload struct {
	EventEmitter  events.EventEmmiter
	RedisClient   *redis.Client
	ElasticClient *elasticsearch.Client
	Logger        *logrus.Logger
}

// ElasticOptions holds configuration values for the elasticsearch cluster
type ElasticOptions struct {
	Endpoint string
	Username string
	Password string
}

//ServerConfiguration holds configuration values for the server
type ServerConfiguration struct {
	Port                    int
	ReadHeaderTimeoutMillis int
}

//GetEnvironment determine the environment in which this application is deployed
func GetEnvironment() string {
	//these will be uppercased automatically
	viper.SetEnvPrefix("env")
	viper.BindEnv("id")

	env := viper.Get("id")

	return fmt.Sprintf("%v", env)
}

//Setup provides application configuration info
func Setup(env string) (*Configuration, error) {
	viper.SetConfigName(env)

	//needed when built at ./cmd/github.com/wambozi/elastic-webcrawler/
	viper.AddConfigPath("../../conf/")
	//needed when built at project root (E.g. when invoked with 'make build')
	viper.AddConfigPath("conf/")
	//needed when unit tests are executed in this package
	viper.AddConfigPath(".")

	var configs Configuration

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file: %w", err)
	}
	err := viper.Unmarshal(&configs)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal into struct: %w", err)
	}

	return &configs, nil
}
