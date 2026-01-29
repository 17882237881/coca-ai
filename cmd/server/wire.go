//go:build wireinject

package main

import (
	"coca-ai/internal/handler"
	"coca-ai/internal/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitApp() *gin.Engine {
	wire.Build(
		handler.NewPingHandler,
		ioc.InitWebServer,
	)
	return nil
}
