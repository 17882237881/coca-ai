package repository

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository/cache"
	"coca-ai/internal/repository/dao"
	"context"
	"time"
)

// MessageRepository 消息仓储层 (支持 Redis 缓存)
type MessageRepository struct {
	dao   *dao.MessageDAO
	cache *cache.MessageCache
}

// NewMessageRepository 创建 MessageRepository 实例
func NewMessageRepository(dao *dao.MessageDAO, cache *cache.MessageCache) *MessageRepository {
	return &MessageRepository{dao: dao, cache: cache}
}

// Create 创建新消息 (Write-Through: 同时写入缓存和数据库)
func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	entity := r.toEntity(message)
	err := r.dao.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	result := r.toDomain(entity)

	// 写入 Redis 缓存
	if r.cache != nil {
		_ = r.cache.Append(ctx, r.toCachedMessage(result))
	}

	return result, nil
}

// BatchCreate 批量创建消息
func (r *MessageRepository) BatchCreate(ctx context.Context, messages []domain.Message) error {
	entities := make([]dao.Message, len(messages))
	for i, msg := range messages {
		entities[i] = *r.toEntity(&msg)
	}
	return r.dao.BatchCreate(ctx, entities)
}

// FindBySessionID 根据会话 ID 查找所有消息 (Read-Through: 优先读缓存)
func (r *MessageRepository) FindBySessionID(ctx context.Context, sessionID int64) ([]domain.Message, error) {
	// 1. 尝试从缓存读取
	if r.cache != nil {
		cached, err := r.cache.GetAll(ctx, sessionID)
		if err == nil && len(cached) > 0 {
			return r.cachedToDomainList(cached), nil
		}
	}

	// 2. 缓存未命中，从数据库读取
	entities, err := r.dao.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, len(entities))
	for i, entity := range entities {
		messages[i] = *r.toDomain(&entity)
	}

	// 3. 回填缓存 (异步，不阻塞主流程)
	if r.cache != nil && len(messages) > 0 {
		go r.warmupCache(context.Background(), sessionID, messages)
	}

	return messages, nil
}

// FindRecentBySessionID 获取会话的最近 N 条消息 (Read-Through)
func (r *MessageRepository) FindRecentBySessionID(ctx context.Context, sessionID int64, limit int) ([]domain.Message, error) {
	// 1. 尝试从缓存读取
	if r.cache != nil {
		cached, err := r.cache.GetRecent(ctx, sessionID, limit)
		if err == nil && len(cached) > 0 {
			return r.cachedToDomainList(cached), nil
		}
	}

	// 2. 缓存未命中，从数据库读取
	entities, err := r.dao.FindRecentBySessionID(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, len(entities))
	for i, entity := range entities {
		messages[i] = *r.toDomain(&entity)
	}

	return messages, nil
}

// CountBySessionID 统计会话的消息数量
func (r *MessageRepository) CountBySessionID(ctx context.Context, sessionID int64) (int64, error) {
	// 优先从缓存获取
	if r.cache != nil {
		count, err := r.cache.GetCount(ctx, sessionID)
		if err == nil && count > 0 {
			return count, nil
		}
	}
	return r.dao.CountBySessionID(ctx, sessionID)
}

// DeleteBySessionID 删除会话的所有消息 (同时删除缓存)
func (r *MessageRepository) DeleteBySessionID(ctx context.Context, sessionID int64) error {
	// 先删除缓存
	if r.cache != nil {
		_ = r.cache.Delete(ctx, sessionID)
	}
	return r.dao.DeleteBySessionID(ctx, sessionID)
}

// AppendToCache 仅写入缓存 (用于异步落库场景)
func (r *MessageRepository) AppendToCache(ctx context.Context, message *domain.Message) error {
	if r.cache == nil {
		return nil
	}
	return r.cache.Append(ctx, r.toCachedMessage(message))
}

// ==================== 私有方法 ====================

// warmupCache 将数据库数据回填到缓存
func (r *MessageRepository) warmupCache(ctx context.Context, sessionID int64, messages []domain.Message) {
	cached := make([]*cache.CachedMessage, len(messages))
	for i, msg := range messages {
		cached[i] = r.toCachedMessage(&msg)
	}
	_ = r.cache.BatchAppend(ctx, cached)
}

// ==================== 转换方法 ====================

// toEntity 将 Domain 转换为 DAO Entity
func (r *MessageRepository) toEntity(message *domain.Message) *dao.Message {
	return &dao.Message{
		Id:        message.ID,
		SessionId: message.SessionID,
		Role:      string(message.Role),
		Content:   message.Content,
		CreatedAt: message.CreatedAt.UnixMilli(),
	}
}

// toDomain 将 DAO Entity 转换为 Domain
func (r *MessageRepository) toDomain(entity *dao.Message) *domain.Message {
	return &domain.Message{
		ID:        entity.Id,
		SessionID: entity.SessionId,
		Role:      domain.MessageRole(entity.Role),
		Content:   entity.Content,
		CreatedAt: time.UnixMilli(entity.CreatedAt),
	}
}

// toCachedMessage 将 Domain 转换为 CachedMessage
func (r *MessageRepository) toCachedMessage(msg *domain.Message) *cache.CachedMessage {
	return &cache.CachedMessage{
		ID:        msg.ID,
		SessionID: msg.SessionID,
		Role:      string(msg.Role),
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.UnixMilli(),
	}
}

// cachedToDomain 将 CachedMessage 转换为 Domain
func (r *MessageRepository) cachedToDomain(cached *cache.CachedMessage) *domain.Message {
	return &domain.Message{
		ID:        cached.ID,
		SessionID: cached.SessionID,
		Role:      domain.MessageRole(cached.Role),
		Content:   cached.Content,
		CreatedAt: time.UnixMilli(cached.CreatedAt),
	}
}

// cachedToDomainList 批量转换
func (r *MessageRepository) cachedToDomainList(cached []*cache.CachedMessage) []domain.Message {
	messages := make([]domain.Message, len(cached))
	for i, c := range cached {
		messages[i] = *r.cachedToDomain(c)
	}
	return messages
}
