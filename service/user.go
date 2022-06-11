package srv

import (
	"context"
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type User struct{}
type UserLoginReq struct {
	Username string `form:"username" binding:"required,min=1,max=32"`
	Password string `form:"password" binding:"required,min=6,max=32"`
}

type UserLoginResp struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

type UserRegisterReq struct {
	Username string `form:"username" binding:"required,min=1,max=32"`
	Password string `form:"password" binding:"required,min=6,max=32"`
}

type UserRegisterResp struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

func (u User) Login(c *gin.Context) (interface{}, error) {
	var req UserLoginReq
	var token string

	// 解析参数
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

	// 密码校验
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Logger.Error("password error", zap.Any("user", user))
		return nil, err
	}

	// acc token: 2h
	token, err = util.CreateAccessToken(user.Id)
	if err != nil {
		log.Logger.Error("create access token error")
		return nil, err
	}

	// ref token 30d
	refreshToken, err := util.CreateRefreshToken(user.Id)
	if err != nil {
		log.Logger.Error("create refresh token error")
		return nil, err
	}

	//c.Header("token", token) // 不需要了

	// key: 2h token; value 30d token; key live time: 30d

	if err := db.Redis.Set(context.Background(), token, refreshToken, 30*24*time.Hour).Err(); err != nil {
		log.Logger.Error("redis set error", zap.Error(err))
		return nil, err
	} else {
		log.Logger.Debug("redis set success")
	}

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
		log.Logger.Error("validate err", zap.Error(err))
		return nil, err
	}

	// 检查是否注册过
	var count int64
	err = db.MySQL.Debug().Model(&model.User{}).Where("username = ?", req.Username).Select("id").Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Logger.Error("mysql happen error")
		return nil, err
	}
	if count != 0 {
		return nil, ErrUsernameExits
	}

	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		Name:     req.Username,
		Username: req.Username,
		Password: string(hash),
	}

	db.MySQL.Debug().Create(&user)

	// acc token: 2h
	token, err := util.CreateAccessToken(user.Id)
	if err != nil {
		log.Logger.Error("create access token error")
		return nil, err
	}

	// ref token 30d
	refreshToken, err := util.CreateRefreshToken(user.Id)
	if err != nil {
		log.Logger.Error("create refresh token error")
		return nil, err
	}

	// key: 2h token; value 30d token; key live time: 30d
	if err := db.Redis.Set(context.Background(), token, refreshToken, 30*24*time.Hour).Err(); err != nil {
		log.Logger.Error("redis set error", zap.Error(err))
		return nil, err
	} else {
		log.Logger.Debug("redis set success")
	}

	return UserRegisterResp{
		UserId: user.Id,
		Token:  token,
	}, nil
}

// Info 依靠用户 ID 查询用户信息，因为还要返回是否关注，所以还要传入当前的用户 ID
// myUserId: get from token; 为0表示请求为传入token
// targetUserId: get from url
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
