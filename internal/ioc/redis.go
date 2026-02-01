package ioc

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	// Default for local development
	addr := "localhost:16379"

	// Override with Docker environment if available
	if envAddr := os.Getenv("REDIS_ADDR"); envAddr != "" {
		addr = envAddr
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}
