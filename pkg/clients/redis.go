package clients

import (
	"github.com/go-redis/redis"
	"github.com/wambozi/elastic-webcrawler/m/conf"
)

// CreateRedisClient returns a new instance of the redis client
func CreateRedisClient(c *conf.RedisOptions) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Host,
		Password: c.Password,
		DB:       c.Database,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
