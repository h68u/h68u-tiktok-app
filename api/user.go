package api

import "github.com/gin-gonic/gin"

type UserHandler interface {

	//查询userid，并由此创建jwt token
	Login(c *gin.Context) (interface{}, error)
}
