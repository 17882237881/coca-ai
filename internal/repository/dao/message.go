package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Message 数据库实体 (对应 messages 表)
type Message struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	SessionId int64  `gorm:"index;not null"`
	Role      string `gorm:"type:varchar(20);not null"` // user, assistant, system
	Content   string `gorm:"type:text;not null"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
}

// TableName 指定表名
func (Message) TableName() string { 
	return "messages"
}

// MessageDAO 消息数据访问对象
type MessageDAO struct {
	db *gorm.DB
}

// NewMessageDAO 创建 MessageDAO 实例
func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

// Create 创建新消息
func (d *MessageDAO) Create(ctx context.Context, message *Message) error {
	message.CreatedAt = time.Now().UnixMilli()
	return d.db.WithContext(ctx).Create(message).Error
}

// BatchCreate 批量创建消息
func (d *MessageDAO) BatchCreate(ctx context.Context, messages []Message) error {
	return d.db.WithContext(ctx).Create(&messages).Error
}

// FindBySessionID 根据会话 ID 查找所有消息，按创建时间正序
func (d *MessageDAO) FindBySessionID(ctx context.Context, sessionID int64) ([]Message, error) {
	var messages []Message
	err := d.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// FindRecentBySessionID 获取会话的最近 N 条消息
func (d *MessageDAO) FindRecentBySessionID(ctx context.Context, sessionID int64, limit int) ([]Message, error) {
	var messages []Message
	err := d.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序，使消息按时间正序排列
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// CountBySessionID 统计会话的消息数量
func (d *MessageDAO) CountBySessionID(ctx context.Context, sessionID int64) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&Message{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return count, err
}

// DeleteBySessionID 删除会话的所有消息
func (d *MessageDAO) DeleteBySessionID(ctx context.Context, sessionID int64) error {
	return d.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&Message{}).Error
}
