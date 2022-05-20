package main

import (
	"github.com/gin-gonic/gin"
	"tikapp/util"
)

func main() {
	r := gin.Default()

	handle(r)

	r.Run(util.GetServerLoc())
}
