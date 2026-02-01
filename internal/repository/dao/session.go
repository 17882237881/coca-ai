package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Session 数据库实体 (对应 sessions 表)
type Session struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	UserId    int64  `gorm:"index"`
	Title     string `gorm:"type:varchar(255);default:''"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli;index"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// SessionDAO 会话数据访问对象
type SessionDAO struct {
	db *gorm.DB
}

// NewSessionDAO 创建 SessionDAO 实例
func NewSessionDAO(db *gorm.DB) *SessionDAO {
	return &SessionDAO{db: db}
}

// Create 创建新会话
func (d *SessionDAO) Create(ctx context.Context, session *Session) error {
	now := time.Now().UnixMilli()
	session.CreatedAt = now
	session.UpdatedAt = now
	return d.db.WithContext(ctx).Create(session).Error
}

// FindByID 根据 ID 查找会话
func (d *SessionDAO) FindByID(ctx context.Context, id int64) (*Session, error) {
	var session Session
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	return &session, err
}

// FindByUserID 根据用户 ID 查找所有会话，按更新时间倒序
func (d *SessionDAO) FindByUserID(ctx context.Context, userID int64) ([]Session, error) {
	var sessions []Session
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateTitle 更新会话标题
func (d *SessionDAO) UpdateTitle(ctx context.Context, id int64, title string) error {
	return d.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", id).
		Update("title", title).Error
}

// Delete 删除会话
func (d *SessionDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&Session{}).Error
}

// TouchUpdatedAt 更新会话的 updated_at 时间戳
func (d *SessionDAO) TouchUpdatedAt(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", id).
		Update("updated_at", time.Now().UnixMilli()).Error
}
