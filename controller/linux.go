package ctrl

import (
	"github.com/gin-gonic/gin"
	res "tikapp/common/result"
	"tikapp/util"
)

func LogToWeb(c *gin.Context) {
	data, err := util.ExecuteCmd("systemctl status app", c)
	if err != nil {
		res.Error(c, res.Status{
			StatusCode: 11001,
			StatusMsg:  "execute cmd error",
		})
	}
	res.Success(c, res.R{
		"log": data,
	})
}
