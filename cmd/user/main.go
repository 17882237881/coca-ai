package main

import (
	userv1 "coca-ai/api/user/v1"
	"coca-ai/internal/ioc"
	"coca-ai/internal/repository"
	"coca-ai/internal/repository/dao"
	"coca-ai/internal/service"
	"coca-ai/internal/service/adapter"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// 1. 初始化依赖 (DB, Redis)
	db := ioc.InitDB()
	rdb := ioc.InitRedis()

	// 2. 初始化 DAO, Repo, Service
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO, rdb)
	userService := service.NewUserService(userRepo)

	// 3. 初始化 gRPC Server
	grpcServer := grpc.NewServer()
	userGRPCServer := adapter.NewUserGRPCServer(userService)
	userv1.RegisterUserServiceServer(grpcServer, userGRPCServer)

	// 4. 监听端口
	addr := ":9090"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("User Service (gRPC) listening on %s\n", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
