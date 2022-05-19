package api

import "github.com/gin-gonic/gin"

type Hello interface {
	/**
	写注释的地方，什么时候会返回error，都可以注明清楚
	*/
	Hello(c *gin.Context) (interface{}, error)
}
