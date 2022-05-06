package main

import (
	"gin_template/hub"
	_ "gin_template/module/pong"
	"gin_template/utils"
	"os"
	"os/signal"
)

func init() {
	utils.WriteLogToFS()
}

func main() {
	hub.Init()

	hub.StartService()

	hub.Run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
}
