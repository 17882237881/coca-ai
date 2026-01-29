package service

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("邮箱或密码错误")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (string, error)
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
	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.repo.Create(ctx, u)
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// 1. 查找用户
	u, err := s.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}

	// 2. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 3. 生成 Token (暂用 Mock，下一步 JWT)
	// TODO: Replace with real JWT implementation
	return "mock_token_" + u.Email, nil
}
