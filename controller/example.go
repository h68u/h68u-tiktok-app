package controller

import (
	"gin_template/service"
	"github.com/gin-gonic/gin"
	"log"
)

/**
对接router的方法
*/
func Hello(c *gin.Context) {
	var e service.Example
	hello, err := e.Hello(c)
	if err != nil {
		log.Fatal(err) //替换通用返回error处理
	}
	c.JSON(200, gin.H{
		"msg": hello,
	}) //替换通用返回success处理
}
