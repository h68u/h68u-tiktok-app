package srv

import (
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type User struct{}
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
		log.Logger.Error("parse json error")
		return nil, err
	}
	var user model.User
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		log.Logger.Error("mysql happen error")
		return nil, err
	}
	// 走不到这？
	//if count != 1 {
	//	log.Logger.Error("no user", zap.Any("count", count))
	//	return nil, err
	//}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Logger.Error("password error", zap.Any("user", user))
		return nil, err
	}
	token, err := util.CreateAccessToken(user.Id)
	if err != nil {
		log.Logger.Error("create access token error")
		return nil, err
	}
	refreshToken, err := util.CreateRefreshToken(user.Id)
	if err != nil {
		log.Logger.Error("create refresh token error")
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
		log.Logger.Error("parse json error")
		log.Logger.Error("validate err", zap.Error(err))
		return nil, err
	}
	// TODO： 这条已经不会运行到，返回的"status_msg"现在都是"register happen error"?
	if len(strings.TrimSpace(req.Username)) == 0 || len(strings.TrimSpace(req.Password)) == 0 {
		return nil, ErrEmpty
	}

	var count int64
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ?", req.Username).Select("id").Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Logger.Error("mysql happen error")
		return nil, err
	}
	if count != 0 {
		return nil, ErrUsernameExits
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		Name:     req.Username,
		Username: req.Username,
		Password: string(hash),
	}

	db.MySQL.Debug().Create(&user)
	token, err := util.CreateAccessToken(user.Id)
	if err != nil {
		log.Logger.Error("create access token error")
		return nil, err
	}
	return UserRegisterResp{
		UserId: user.Id,
		Token:  token,
	}, nil
}

// Info 依靠用户 ID 查询用户信息，因为还要返回是否关注，所以还要传入当前的用户 ID
func (u User) Info(myUserID, targetUserID int64) (UserDemo, error) {
	var userInTable model.User //返回的格式和表中格式不一样
	var user UserDemo
	var followCount int64

	// 查询用户信息
	err := db.MySQL.Debug().Model(&model.User{}).Where("id = ?", targetUserID).First(&userInTable).Error
	if err != nil {
		log.Logger.Error("mysql happen error")
		return UserDemo{}, err
	}

	// 把临时结构体中的信息拷贝至返回体中
	user.Id = userInTable.Id
	user.Name = userInTable.Name
	user.FollowerCount = userInTable.FollowerCount
	user.FollowCount = userInTable.FollowCount

	// 检查是否关注
	if myUserID == 0 || myUserID == targetUserID {
		return user, nil // 游客和查看自己的主页自然没有关注这一说，直接返回
	}
	err = db.MySQL.Debug().Model(&model.Follow{}).Where("follow_id = ? and user_id = ?", myUserID, targetUserID).Count(&followCount).Error
	if err != nil {
		log.Logger.Error("mysql happen error")
		return UserDemo{}, err
	}

	if followCount != 0 {
		user.IsFollow = true
	}

	return user, nil
}
