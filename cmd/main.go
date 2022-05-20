package main

import (

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	handle(r)

	r.Run(":8080")
}
	

