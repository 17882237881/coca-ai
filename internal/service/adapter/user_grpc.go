package adapter

import (
	userv1 "coca-ai/api/user/v1"
	"coca-ai/internal/domain"
	"coca-ai/internal/service"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserGRPCServer 适配层，将 gRPC 请求转发给 UseCase (service.UserService)
type UserGRPCServer struct {
	userv1.UnimplementedUserServiceServer
	svc service.UserService
}

func NewUserGRPCServer(svc service.UserService) *UserGRPCServer {
	return &UserGRPCServer{
		svc: svc,
	}
}

func (s *UserGRPCServer) Signup(ctx context.Context, req *userv1.SignupReq) (*userv1.SignupResp, error) {
	if req.Password != req.ConfirmPassword {
		return nil, status.Error(codes.InvalidArgument, "密码不一致")
	}

	u := domain.User{
		Email:    req.Email,
		Password: req.Password,
	}

	if err := s.svc.Signup(ctx, u); err != nil {
		if errors.Is(err, service.ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, "邮箱已注册")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 这里的 id 获取逻辑可能需要调整，因为 Signup 接口当前不返回 ID
	// 暂时假设 repo 会在 Signup 里处理好或我们后续通过 Login 获取
	// 简单起见，这里返回 0 或修改 svc 接口
	return &userv1.SignupResp{}, nil
}

func (s *UserGRPCServer) Login(ctx context.Context, req *userv1.LoginReq) (*userv1.LoginResp, error) {
	id, ssid, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.LoginResp{
		Id:   id,
		Ssid: ssid,
	}, nil
}

func (s *UserGRPCServer) Logout(ctx context.Context, req *userv1.LogoutReq) (*userv1.LogoutResp, error) {
	if err := s.svc.Logout(ctx, req.Ssid); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userv1.LogoutResp{}, nil
}
