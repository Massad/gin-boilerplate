package db

import (
	"fmt"
	"log"
	"os"
	"strings"

	_redis "github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var dbConn *sqlx.DB

// Init connects to PostgreSQL using environment variables.
func Init() {
	var err error
	dbConn, err = sqlx.Connect("postgres", postgresDSN(os.Getenv))
	if err != nil {
		log.Fatal(err)
	}
}

func postgresDSN(getenv func(string) string) string {
	sslMode := "disable"
	if getenv("SSL") == "TRUE" {
		sslMode = "require"
	}

	// lib/pq keyword values must be quoted, including empty strings. Otherwise
	// whitespace or quotes in a value can change how later options are parsed.
	quote := func(value string) string {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		quote(getenv("DB_HOST")),
		quote(getenv("DB_PORT")),
		quote(getenv("DB_USER")),
		quote(getenv("DB_PASS")),
		quote(getenv("DB_NAME")),
		sslMode,
	)
}

// GetDB returns the sqlx database connection.
func GetDB() *sqlx.DB {
	return dbConn
}

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
