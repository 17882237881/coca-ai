package ioc

import (
	"coca-ai/internal/handler"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitWebServer(pingHandler *handler.PingHandler) *gin.Engine {
	server := gin.Default()

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
	return server
}
