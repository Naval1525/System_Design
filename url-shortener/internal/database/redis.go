package database

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

const (
	defaultRedisHost = "localhost"
	defaultRedisPort = "6379"
)

func NewRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = defaultRedisHost
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = defaultRedisPort
	}
	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return rdb
}

// RedisAddr returns the address used by NewRedisClient (same env logic).
// Use for logging so you can confirm which Redis the app talks to.
func RedisAddr() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = defaultRedisHost
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = defaultRedisPort
	}
	return fmt.Sprintf("%s:%s", host, port)
}
