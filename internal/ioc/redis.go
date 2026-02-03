package ioc

import (
	"coca-ai/internal/config"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	// Default for local development
	cfg := config.Get()
	addr := cfg.Redis.Addr
	if addr == "" {
		addr = "localhost:16379"
	}

	// Override with Docker environment if available
	if envAddr := os.Getenv("REDIS_ADDR"); envAddr != "" {
		addr = envAddr
	}

	password := cfg.Redis.Password
	if envPassword := os.Getenv("REDIS_PASSWORD"); envPassword != "" {
		password = envPassword
	}

	db := cfg.Redis.DB
	if envDB := os.Getenv("REDIS_DB"); envDB != "" {
		if v, err := strconv.Atoi(envDB); err == nil {
			db = v
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     orDefaultInt(cfg.Redis.PoolSize, 50), // 设置连接池大小
		MinIdleConns: orDefaultInt(cfg.Redis.MinIdleConns, 10), // 设置最小空闲连接数
		DialTimeout:  orDefaultDuration(cfg.Redis.DialTimeoutMS, 3000), // 设置拨号超时时间
		ReadTimeout:  orDefaultDuration(cfg.Redis.ReadTimeoutMS, 3000), // 设置读取超时时间
		WriteTimeout: orDefaultDuration(cfg.Redis.WriteTimeoutMS, 3000), // 设置写入超时时间
	})
	return client
}

func orDefaultInt(value int, def int) int {
	if value > 0 {
		return value
	}
	return def
}

func orDefaultDuration(ms int, defMs int) time.Duration {
	if ms <= 0 {
		ms = defMs
	}
	return time.Duration(ms) * time.Millisecond
}
