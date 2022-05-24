package ctrl

import "github.com/gin-gonic/gin"

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{"status": "测试webhook1!"})
}
