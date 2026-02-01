package main

import (
	"coca-ai/internal/ioc"
)

func main() {
	ioc.InitJaeger()
	server := InitApp()
	server.Run(":8080")
}
