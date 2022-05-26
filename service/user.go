package srv

import (
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
	"tikapp/api"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// 这个貌似没有用
var userLogger = log.NameSpace("UserService")

type User struct{}

// 根据 Uber 的指导原则 这里是检查 User 是否实现了 api 中的所有方法
// 即检查项目是否缺少必要的接口
var _ api.UserHandler = (*User)(nil)

type UserLoginReq struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

type UserLoginResp struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

type UserRegisterReq struct {
	Username string `form:"username" binding:"required,min=1,max=32"`
	Password string `form:"password" binding:"required,min=5,max=32"`
}

type UserRegisterResp struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

func (u User) Login(c *gin.Context) (interface{}, error) {
	var req UserLoginReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		userLogger.Error("parse json error")
		return nil, err
	}
	var user model.User
	var count int64
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ? and password = ?", req.Username, req.Password).Select("id").First(&user).Count(&count).Error
	if err != nil {
		userLogger.Error("mysql happen error")
		return nil, err
	}
	if count != 1 {
		userLogger.Error("no user or user repeat")
		return nil, err
	}
	token, err := util.CreateAccessToken(user.Id)
	if err != nil {
		userLogger.Error("create access token error")
		return nil, err
	}
	refreshToken, err := util.CreateRefreshToken(user.Id)
	if err != nil {
		userLogger.Error("create refresh token error")
		return nil, err
	}
	c.Header("token", token)
	db.Redis.Set(token, refreshToken, 30*24*time.Hour)
	return UserLoginResp{
		UserId: user.Id,
		Token:  token,
	}, nil
}

var ErrUsernameExits = errors.New("username already exists")
var ErrEmpty = errors.New("username or password is empty")

func (u User) Register(c *gin.Context) (interface{}, error) {
	var req UserRegisterReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		userLogger.Error("parse json error")
		log.Logger.Error("validate err", zap.Error(err))
		return nil, err
	}

	if len(strings.TrimSpace(req.Username)) == 0 || len(strings.TrimSpace(req.Password)) == 0 {
		return nil, ErrEmpty
	}

	var count int64
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ?", req.Username).Select("id").Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		userLogger.Error("mysql happen error")
		return nil, err
	}
	if count != 0 {
		return nil, ErrUsernameExits
	}
	user := model.User{
		Name:     req.Username,
		Username: req.Username,
		Password: req.Password,
	}

	db.MySQL.Debug().Create(&user)
	token, err := util.CreateAccessToken(user.Id)
	if err != nil {
		userLogger.Error("create access token error")
		return nil, err
	}
	return UserRegisterResp{
		UserId: user.Id,
		Token:  token,
	}, nil
}

// Info 依靠用户 ID 查询用户信息，因为还要返回是否关注，所以还要传入当前的用户 ID
func (u User) Info(myUserID, targetUserID int64) (model.User, bool, error) {
	var user model.User
	var isFollow int64

	// 查询用户信息
	err := db.MySQL.Debug().Model(&model.User{}).Where("id = ?", targetUserID).First(&user).Error
	if err != nil {
		userLogger.Error("mysql happen error")
		return model.User{}, false, err
	}

	// 检查是否关注
	if myUserID == 0 || myUserID == targetUserID {
		return user, false, nil // 游客和查看自己的主页自然没有关注这一说，直接返回
	}
	err = db.MySQL.Debug().Model(&model.Follow{}).Where("follow_id = ? and user_id = ?", myUserID, targetUserID).Count(&isFollow).Error
	if err != nil {
		userLogger.Error("mysql happen error")
		return model.User{}, false, err
	}

	return user, isFollow > 0, nil
}
