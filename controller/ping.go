package ctrl

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	url := c.Request.URL
	fmt.Println(url.String())
	uri := c.Request.RequestURI
	fmt.Println(uri)
	c.JSON(200, gin.H{"status": "测试webhook!"})
}
