package srv

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
)

var logger = log.NameSpace("UserService")

type User struct{}

type UserLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginResp struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

func (u User) Login(c *gin.Context) (interface{}, error) {
	var req UserLoginReq
	err := c.ShouldBindWith(&req, binding.JSON)
	if err != nil {
		logger.Error("parse json error")
		return nil, err
	}
	var user model.User
	var count int64
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ? and password = ?", req.Username, req.Password).Select("id").First(&user).Count(&count).Error
	if err != nil {
		logger.Error("mysql happen error")
		return nil, err
	}
	if count != 1 {
		logger.Error("no user or user repeat")
		return nil, err
	}
	token, err := util.GenerateToken(user.Id)
	if err != nil {
		logger.Error("create token error")
		return nil, err
	}
	//此处待决定
	c.SetCookie("token", token, 30*24*60*60*1000, "/", "localhost", false, true)
	return UserLoginResp{
		UserId: user.Id,
		Token:  token,
	}, nil
}
