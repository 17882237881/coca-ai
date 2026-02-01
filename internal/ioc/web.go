package ioc

import (
	"coca-ai/internal/handler"
	"coca-ai/internal/handler/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitWebServer(pingHandler *handler.PingHandler, userHandler *handler.UserHandler, chatHandler *handler.ChatHandler, jwtMiddleware *middleware.LoginJWTMiddleware) *gin.Engine {
	server := gin.Default()

	// 初始化 Prometheus 监控
	InitPrometheus(server)

	// CORS Configuration
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // For dev allow all
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	pingHandler.RegisterRoutes(server)
	userHandler.RegisterRoutes(server, jwtMiddleware)
	chatHandler.RegisterRoutes(server, jwtMiddleware.Check())
	return server
}
