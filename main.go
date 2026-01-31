package main

import (
	"log"

	"jijin/internal/app"
)

func main() {
	application := app.NewApp()
	if err := application.Run(); err != nil {
		log.Fatal("启动失败:", err)
	}
}
