package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

// Configuration holds configuration values for the application
type Configuration struct {
	Server        ServerConfiguration
	Elasticsearch ElasticOptions
	Appsearch     AppsearchOptions
}

// ElasticOptions holds configuration values for the elasticsearch cluster
type ElasticOptions struct {
	Endpoint string
	Username string
	Password string
}

// AppsearchOptions holds config values for the app-search instance
type AppsearchOptions struct {
	Endpoint string
	API      string
	Token    string
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

	if env == "local" {
		viper.SetConfigName("local")
	}

	// used for unit testing
	if env == "test" {
		viper.SetConfigName("test")
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
