package ioc

import (
	"coca-ai/internal/config"
	"coca-ai/internal/repository/cache"

	"github.com/redis/go-redis/v9"
)

// InitMessageCache 初始化消息缓存
func InitMessageCache(cmd redis.Cmdable) *cache.MessageCache {
	cfg := config.Get()
	maxLen := int64(cfg.Redis.MessageCacheMaxLen)
	if maxLen < 0 {
		maxLen = 0
	}
	return cache.NewMessageCache(cmd, maxLen)
}
