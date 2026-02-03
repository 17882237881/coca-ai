package ioc

import (
	userv1 "coca-ai/api/user/v1"
	"coca-ai/internal/service"
	"coca-ai/internal/service/adapter"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitUserGRPCClient() service.UserService {
	addr := os.Getenv("USER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:9090" // Default to local for dev
	}
	// 建立连接
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	// 创建客户端
	client := userv1.NewUserServiceClient(conn)

	// 包装为 Service 接口
	return adapter.NewUserGRPCClient(client)
}
