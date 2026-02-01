//go:build wireinject
// +build wireinject

package main

import (
	"coca-ai/internal/handler"
	"coca-ai/internal/handler/middleware"
	"coca-ai/internal/ioc"
	"coca-ai/internal/repository"
	"coca-ai/internal/repository/cache"
	"coca-ai/internal/repository/dao"
	"coca-ai/internal/service"
	"coca-ai/pkg/jwtx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitApp() *gin.Engine {
	wire.Build(
		// 基础组件
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitWebServer,
		// LLM 客户端
		ioc.InitLLMClient,
		// Kafka
		ioc.InitKafkaProducer,
		// User 模块
		dao.NewUserDAO,
		repository.NewUserRepository,
		service.NewUserService,
		jwtx.NewJWTHandler,
		handler.NewUserHandler,
		middleware.NewLoginJWTMiddleware,
		handler.NewPingHandler,
		// Chat 模块
		dao.NewSessionDAO,
		dao.NewMessageDAO,
		cache.NewMessageCache,
		repository.NewSessionRepository,
		repository.NewMessageRepository,
		service.NewContextService,
		service.NewChatService,
		handler.NewChatHandler,
	)
	return nil
}
