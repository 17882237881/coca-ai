package repository

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository/dao"
	"context"
	"time"
)

// SessionRepository 会话仓储层
type SessionRepository struct {
	dao *dao.SessionDAO
}

// NewSessionRepository 创建 SessionRepository 实例
func NewSessionRepository(dao *dao.SessionDAO) *SessionRepository {
	return &SessionRepository{dao: dao}
}

// Create 创建新会话
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	entity := r.toEntity(session)
	err := r.dao.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	return r.toDomain(entity), nil
}

// FindByID 根据 ID 查找会话
func (r *SessionRepository) FindByID(ctx context.Context, id int64) (*domain.Session, error) {
	entity, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(entity), nil
}

// FindByUserID 根据用户 ID 查找所有会话
func (r *SessionRepository) FindByUserID(ctx context.Context, userID int64) ([]domain.Session, error) {
	entities, err := r.dao.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sessions := make([]domain.Session, len(entities))
	for i, entity := range entities {
		sessions[i] = *r.toDomain(&entity)
	}
	return sessions, nil
}

// UpdateTitle 更新会话标题
func (r *SessionRepository) UpdateTitle(ctx context.Context, id int64, title string) error {
	return r.dao.UpdateTitle(ctx, id, title)
}

// Delete 删除会话
func (r *SessionRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

// TouchUpdatedAt 更新会话的更新时间
func (r *SessionRepository) TouchUpdatedAt(ctx context.Context, id int64) error {
	return r.dao.TouchUpdatedAt(ctx, id)
}

// ==================== 转换方法 ====================

// toEntity 将 Domain 转换为 DAO Entity
func (r *SessionRepository) toEntity(session *domain.Session) *dao.Session {
	return &dao.Session{
		Id:        session.ID,
		UserId:    session.UserID,
		Title:     session.Title,
		CreatedAt: session.CreatedAt.UnixMilli(),
		UpdatedAt: session.UpdatedAt.UnixMilli(),
	}
}

// toDomain 将 DAO Entity 转换为 Domain
func (r *SessionRepository) toDomain(entity *dao.Session) *domain.Session {
	return &domain.Session{
		ID:        entity.Id,
		UserID:    entity.UserId,
		Title:     entity.Title,
		CreatedAt: time.UnixMilli(entity.CreatedAt),
		UpdatedAt: time.UnixMilli(entity.UpdatedAt),
	}
}
