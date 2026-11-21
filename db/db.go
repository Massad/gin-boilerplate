package db

import (
	"os"

	_redis "github.com/go-redis/redis/v7"
)

// RedisClient holds the Redis connection.
var RedisClient *_redis.Client

// InitRedis connects to Redis.
func InitRedis(selectDB ...int) {
	var redisHost = os.Getenv("REDIS_HOST")
	var redisPassword = os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if len(selectDB) > 0 {
		redisDB = selectDB[0]
	}

	RedisClient = _redis.NewClient(&_redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       redisDB,
	})
}

// GetRedis returns the Redis client.
func GetRedis() *_redis.Client {
	return RedisClient
}
