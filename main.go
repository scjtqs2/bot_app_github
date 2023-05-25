// Package main 入口
package main

import (
	"os"
	"os/signal"

	"github.com/scjtqs2/bot_app_github/app"
)

func main() {
	newApp := app.NewApp()
	newApp.Init()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	os.Exit(1)
}
