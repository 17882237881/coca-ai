package main

import (
	"context"
	"coca-ai/internal/mq"

	"github.com/gin-gonic/gin"
)

type App struct {
	Engine   *gin.Engine
	Consumer *mq.Consumer
}

func NewApp(engine *gin.Engine, consumer *mq.Consumer) *App {
	return &App{
		Engine:   engine,
		Consumer: consumer,
	}
}

func (a *App) Run(addr string) error {
	if a.Consumer != nil {
		a.Consumer.StartAsync(context.Background())
	}
	return a.Engine.Run(addr)
}
