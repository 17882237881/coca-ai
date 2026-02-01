package service

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail     = repository.ErrDuplicateEmail
	ErrInvalidCredentials = errors.New("用户名或密码错误")
)

type UserService interface {
	// Signup 注册
	Signup(ctx context.Context, u domain.User) error
	// Login 登录
	Login(ctx context.Context, email, password string) (int64, string, error)
	// Logout 登出
	Logout(ctx context.Context, ssid string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Signup(ctx context.Context, u domain.User) error {
	// 1. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	// 2. 存储
	return s.repo.Create(ctx, u)
}

func (s *userService) Login(ctx context.Context, email, password string) (int64, string, error) {
	// 1. 找用户
	u, err := s.repo.FindByEmail(ctx, email)
	switch err {
	case repository.ErrUserNotFound:
		return 0, "", ErrInvalidCredentials
	case nil:
		// 验证密码
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		if err != nil {
			return 0, "", ErrInvalidCredentials
		}
		// 生成 SSID (UUID)
		ssid := uuid.New().String()
		return u.Id, ssid, nil
	default:
		return 0, "", err
	}
}

func (s *userService) Logout(ctx context.Context, ssid string) error {
	// 7天过期 (默认)
	return s.repo.BlockSSID(ctx, ssid, time.Hour*24*7)
}
