package redisdb

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	_redis "github.com/go-redis/redis/v7"
)

// RedisClient holds the Redis connection.
var RedisClient *_redis.Client

// DefaultRedisDB is the Redis index used when REDIS_DB is absent or empty:
// the dev sessions. Tests and the e2e cycle pin their own disposable indexes.
const DefaultRedisDB = 1

// RedisDBIndex resolves the Redis index from REDIS_DB: an absent or empty
// value means DefaultRedisDB, anything else must be an integer within the
// default Redis range 0-15 (`databases 16` is the stock setting of the
// dev/CI containers; the server's `databases` directive remains
// authoritative).
func RedisDBIndex(getenv func(string) string) (int, error) {
	raw := strings.TrimSpace(getenv("REDIS_DB"))
	if raw == "" {
		return DefaultRedisDB, nil
	}
	index, err := strconv.Atoi(raw)
	if err != nil || index < 0 || index > 15 {
		return 0, fmt.Errorf("error: REDIS_DB %q must be an integer between 0 and 15", raw)
	}
	return index, nil
}

// InitRedis connects to Redis on the REDIS_DB index (default DefaultRedisDB).
func InitRedis() error {
	index, err := RedisDBIndex(os.Getenv)
	if err != nil {
		return err
	}
	RedisClient = _redis.NewClient(&_redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       index,
	})
	return nil
}

// GetRedis returns the Redis client.
func GetRedis() *_redis.Client {
	return RedisClient
}
