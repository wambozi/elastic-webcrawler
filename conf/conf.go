package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

//Configuration holds configuration values for the application
type Configuration struct {
	Server        ServerConfiguration
	Elasticsearch ElasticsearchConfiguration
	Redis         RedisConfiguration
	AWS           AwsConfig
}

//AwsConfig represents the values required to instantiate AWS services/clients
type AwsConfig struct {
	Region string
}

//ElasticsearchConfiguration holds configuration values for the Elasticsearch cluster
type ElasticsearchConfiguration struct {
	SecretName string
}

//RedisConfiguration holds configuration values for the data Redis cache
type RedisConfiguration struct {
	SecretName string
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
	//default config name (that does not exist) to intentionally cause errors on startup if config file not found
	viper.SetConfigName("no-config-set")

	if env == "lle" {
		viper.SetConfigName("lower-level")
	}

	if env == "prod" {
		viper.SetConfigName("prod")
	}

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
