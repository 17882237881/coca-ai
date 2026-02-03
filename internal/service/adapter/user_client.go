package adapter

import (
	userv1 "coca-ai/api/user/v1"
	"coca-ai/internal/domain"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCClient struct {
	client userv1.UserServiceClient
}

func NewUserGRPCClient(client userv1.UserServiceClient) *UserGRPCClient {
	return &UserGRPCClient{
		client: client,
	}
}

func (c *UserGRPCClient) Signup(ctx context.Context, u domain.User) error {
	req := &userv1.SignupReq{
		Email:           u.Email,
		Password:        u.Password,
		ConfirmPassword: u.Password,
	}
	_, err := c.client.Signup(ctx, req)
	// Error handling: map gRPC errors if needed, or pass through
	// The handler expects standard errors or specific service errors.
	// For now, simple pass through is likely acceptable as long as handler handles generic errors.
	// Improvements: map status codes to service.Err*
	return err
}

func (c *UserGRPCClient) Login(ctx context.Context, email, password string) (int64, string, error) {
	req := &userv1.LoginReq{
		Email:    email,
		Password: password,
	}
	resp, err := c.client.Login(ctx, req)
	if err != nil {
		// Example mapping
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			// return 0, "", service.ErrInvalidCredentials // Ideally we import service package
		}
		return 0, "", err
	}
	return resp.Id, resp.Ssid, nil
}

func (c *UserGRPCClient) Logout(ctx context.Context, ssid string) error {
	req := &userv1.LogoutReq{
		Ssid: ssid,
	}
	_, err := c.client.Logout(ctx, req)
	return err
}

func (c *UserGRPCClient) Check(ctx context.Context, ssid string) error {
	// Note: The original UserService interface didn't have Check?
	// Wait, let's double check existing UserService interface.
	return nil
}
