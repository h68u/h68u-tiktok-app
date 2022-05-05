package main

import (
	_ "gin_template/module/pong"
	"gin_template/server"
	"gin_template/utils"
	"os"
	"os/signal"
)

func init() {
	utils.WriteLogToFS()
}

func main() {
	server.Init()

	server.StartService()

	server.Run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	server.Stop()
}
