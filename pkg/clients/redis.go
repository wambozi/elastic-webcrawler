package clients

import (
	"github.com/go-redis/redis"
	"github.com/kataras/go-events"
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

// AddLinkToSet adds a link to a given key in redis
func AddLinkToSet(c *redis.Client, e events.EventEmmiter, key string, links []string) int64 {
	res := c.SAdd(key, links)
	if key == "READY" {
		e.Emit("READY", links)
	}
	return res.Val()
}

// GetLinkFromSet gets a link (at random) from a given key in redis
func GetLinkFromSet(c *redis.Client, e events.EventEmmiter, key string) string {
	res := c.SPop(key)
	return res.Val()
}

// ClearSet deletes a given key from Redis
func ClearSet(c *redis.Client, e events.EventEmmiter, key string) int64 {
	res := c.Del(key)
	return res.Val()
}
