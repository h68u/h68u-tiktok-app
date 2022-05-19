package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Example struct {
}

func (e Example) Hello(c *gin.Context) (interface{}, error) {
	n, err := fmt.Println("hello world")
	if err != nil {
		return nil, err
	}
	return n, nil
}
