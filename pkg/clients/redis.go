package clients

import (
	"log"

	"github.com/go-redis/redis"
	"github.com/kataras/go-events"
)

// RedisOptions for the Redis Client
type RedisOptions struct {
	Host     string
	Password string
	Database int
}

// CreateRedisClient returns a new instance of the redis client
func CreateRedisClient(c *RedisOptions) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Host,
		Password: c.Password,
		DB:       c.Database,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	if pong != "" {
		log.Println("Creating Redis Client... Done.")
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
