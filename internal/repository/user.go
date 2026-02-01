package repository

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository/dao"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository interface {
	// FindByEmail 查找用户
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	// FindById 查找用户
	FindById(ctx context.Context, id int64) (domain.User, error)
	// Create 创建用户
	Create(ctx context.Context, u domain.User) error
	// BlockSSID 封禁 SSID
	BlockSSID(ctx context.Context, ssid string, expiration time.Duration) error
}

type userRepository struct {
	dao      dao.UserDAO
	redisCmd redis.Cmdable 
}

func NewUserRepository(dao dao.UserDAO, redisCmd redis.Cmdable) UserRepository {
	return &userRepository{
		dao:      dao,
		redisCmd: redisCmd,
	}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *userRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *userRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.toEntity(u))
}

func (r *userRepository) BlockSSID(ctx context.Context, ssid string, expiration time.Duration) error {
	key := "users:ssid:" + ssid
	return r.redisCmd.Set(ctx, key, "", expiration).Err()
}

func (r *userRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (r *userRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
