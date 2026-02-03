package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 消息缓存 Key 前缀: chat:session:{session_id}:messages
	messageKeyPrefix = "chat:session:%d:messages"
	// 缓存过期时间: 24 小时
	messageTTL = 24 * time.Hour
)

// CachedMessage Redis 中缓存的消息结构
type CachedMessage struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"` // Unix 毫秒
}

// MessageCache 消息缓存操作封装
type MessageCache struct {
	client redis.Cmdable
	maxLen int64
}

// NewMessageCache 创建 MessageCache 实例
func NewMessageCache(client redis.Cmdable, maxLen int64) *MessageCache {
	return &MessageCache{client: client, maxLen: maxLen}
}

// ==================== 写入操作 ====================

// Append 追加一条消息到会话缓存 (RPUSH)
func (c *MessageCache) Append(ctx context.Context, msg *CachedMessage) error {
	key := c.buildKey(msg.SessionID)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	// RPUSH 添加到列表尾部
	pipe := c.client.TxPipeline()
	pipe.RPush(ctx, key, data)
	pipe.Expire(ctx, key, messageTTL)
	if c.maxLen > 0 {
		pipe.LTrim(ctx, key, -c.maxLen, -1)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline failed: %w", err)
	}

	return nil
}

// BatchAppend 批量追加消息
func (c *MessageCache) BatchAppend(ctx context.Context, messages []*CachedMessage) error {
	if len(messages) == 0 {
		return nil
	}

	sessionID := messages[0].SessionID
	key := c.buildKey(sessionID)

	// 序列化所有消息
	values := make([]interface{}, len(messages))
	for i, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal message failed: %w", err)
		}
		values[i] = data
	}

	// 批量 RPUSH
	pipe := c.client.TxPipeline()
	pipe.RPush(ctx, key, values...)
	pipe.Expire(ctx, key, messageTTL)
	if c.maxLen > 0 {
		pipe.LTrim(ctx, key, -c.maxLen, -1)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline failed: %w", err)
	}

	return nil
}

// ==================== 读取操作 ====================

// GetAll 获取会话的所有缓存消息 (LRANGE 0 -1)
func (c *MessageCache) GetAll(ctx context.Context, sessionID int64) ([]*CachedMessage, error) {
	key := c.buildKey(sessionID)

	data, err := c.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis LRANGE failed: %w", err)
	}

	return c.parseMessages(data)
}

// GetRecent 获取会话的最近 N 条消息 (LRANGE -N -1)
func (c *MessageCache) GetRecent(ctx context.Context, sessionID int64, limit int) ([]*CachedMessage, error) {
	key := c.buildKey(sessionID)

	// LRANGE -limit -1 获取最后 limit 条
	data, err := c.client.LRange(ctx, key, int64(-limit), -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis LRANGE failed: %w", err)
	}

	return c.parseMessages(data)
}

// GetCount 获取会话的缓存消息数量 (LLEN)
func (c *MessageCache) GetCount(ctx context.Context, sessionID int64) (int64, error) {
	key := c.buildKey(sessionID)
	return c.client.LLen(ctx, key).Result()
}

// Exists 检查会话是否有缓存消息
func (c *MessageCache) Exists(ctx context.Context, sessionID int64) (bool, error) {
	key := c.buildKey(sessionID)
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== 删除操作 ====================

// Delete 删除会话的所有缓存消息
func (c *MessageCache) Delete(ctx context.Context, sessionID int64) error {
	key := c.buildKey(sessionID)
	return c.client.Del(ctx, key).Err()
}

// ==================== 辅助方法 ====================

// buildKey 构建 Redis Key
func (c *MessageCache) buildKey(sessionID int64) string {
	return fmt.Sprintf(messageKeyPrefix, sessionID)
}

// parseMessages 解析消息列表
func (c *MessageCache) parseMessages(data []string) ([]*CachedMessage, error) {
	messages := make([]*CachedMessage, 0, len(data))
	for _, item := range data {
		var msg CachedMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			return nil, fmt.Errorf("unmarshal message failed: %w", err)
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

// Refresh 刷新缓存过期时间
func (c *MessageCache) Refresh(ctx context.Context, sessionID int64) error {
	key := c.buildKey(sessionID) 
	return c.client.Expire(ctx, key, messageTTL).Err()
}
