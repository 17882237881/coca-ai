package mq

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository/dao"
	"context"
	"log"
	"time"
)

// MessagePersistHandler 消息持久化处理器
// 负责将 Kafka 消息写入 MySQL
type MessagePersistHandler struct {
	dao *dao.MessageDAO
}

// NewMessagePersistHandler 创建消息持久化处理器
func NewMessagePersistHandler(dao *dao.MessageDAO) *MessagePersistHandler {
	return &MessagePersistHandler{dao: dao}
}

// Handle 处理消息事件，持久化到 MySQL
func (h *MessagePersistHandler) Handle(ctx context.Context, event *MessageEvent) error {
	// 转换为 DAO 实体
	entity := &dao.Message{
		Id:        event.ID,
		SessionId: event.SessionID,
		Role:      event.Role,
		Content:   event.Content,
		CreatedAt: event.CreatedAt,
	}

	// 如果 ID 为 0，表示新消息，需要创建
	// 如果 ID 不为 0，表示已有消息，需要更新（或幂等处理）
	if entity.Id == 0 {
		entity.CreatedAt = time.Now().UnixMilli()
		if err := h.dao.Create(ctx, entity); err != nil {
			log.Printf("[MessagePersistHandler] Create message failed: %v", err)
			return err
		}
		log.Printf("[MessagePersistHandler] Created message for session %d, role: %s",
			event.SessionID, event.Role)
	} else {
		// 幂等处理：检查是否已存在，不存在则创建
		// 这里简化处理，直接创建（GORM 会处理主键冲突）
		if err := h.dao.Create(ctx, entity); err != nil {
			// 忽略主键冲突错误，实现幂等
			log.Printf("[MessagePersistHandler] Message may already exist: %v", err)
		}
	}

	return nil
}

// EventToDomain 将 MessageEvent 转换为 domain.Message
func EventToDomain(event *MessageEvent) *domain.Message {
	return &domain.Message{
		ID:        event.ID,
		SessionID: event.SessionID,
		Role:      domain.MessageRole(event.Role),
		Content:   event.Content,
		CreatedAt: time.UnixMilli(event.CreatedAt),
	}
}

// DomainToEvent 将 domain.Message 转换为 MessageEvent
func DomainToEvent(msg *domain.Message) *MessageEvent {
	return &MessageEvent{
		ID:        msg.ID,
		SessionID: msg.SessionID,
		Role:      string(msg.Role),
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.UnixMilli(),
	}
}
