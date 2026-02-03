package main

import (
	"coca-ai/internal/ioc"
	"log"
)

func main() {
	ioc.InitJaeger()
	app := InitApp()
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
