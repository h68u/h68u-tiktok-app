package ctrl

import (
	"github.com/gin-gonic/gin"
	res "tikapp/common/result"
	srv "tikapp/service"
)

func Login(c *gin.Context) {
	var u srv.User
	login, err := u.Login(c)
	if err != nil {
		res.Error(c, res.Status{
			StatusCode: res.LoginErrorStatus.StatusCode,
			StatusMsg:  res.LoginErrorStatus.StatusMsg,
		})
		return
	}
	data := login.(srv.UserLoginResp)
	res.Success(c, res.R{
		"userid": data.UserId,
		"token":  data.Token,
	})
}
