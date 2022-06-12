package middlewire

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/util"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Auth 鉴权接口
// 所有接口使用acc token(auth 2h)
// 在Redis中 acc token(2h)为key，ref token(30d)为value
//
// auth 核心过程：
// 1.检查这个acc token，若没过期取uid
// (已解决小问题：res token 可能过期；比如acc token才生成，res token 还剩1h 过期——这种情况不会出现，acc 与 ref 同时更新)
//
// 2.若acc token过期了，取出ref token
// 检查 ref token是否过期，过期叫重新登陆；
// ref token没过期，生成新的acc token，ref token(防止恰好此时失效，同时更新)，删除旧的记录，新建Redis记录
//
// 3.acc和ref同时更新，只有在30天没有没有登录时提醒重新登陆
//

type BackendLoginReq struct {
}

func Auth() gin.HandlerFunc {
	//先判断请求头是否为空，为空则为游客状态
	return func(c *gin.Context) {
		token := ""
		token = c.DefaultQuery("token", "")
		if token == "" {
			token = c.PostForm("token")
		}
		if token == "" {
			c.Set("userId", "")
			c.Next()
			return
		}

		//有token，判断是否过期: 2h
		timeOut, err := util.ValidToken(token)

		if err != nil || timeOut {
			//token过期或者解析token发生错误
			log.Logger.Info("token expire or parse token error")
			log.Logger.Debug("valid token err", zap.Error(err))
			log.Logger.Info("valid refreshToken")

			// 30d token 能否取出
			value := db.Redis.Get(context.Background(), token)
			refreshToken, err := value.Result()
			if err != nil {
				// debug
				log.Logger.Debug("get refreshToken from redis err", zap.Error(err))

				log.Logger.Error("token不合法，请确认你是否登录")
				c.JSON(200, gin.H{
					"status_code": 400,
					"msg":         "token不合法，请确认你操作是否有误",
				})
				c.Abort()
				return
			}

			// 可以取出30d token, 检查是否过期
			timeOut, err := util.ValidToken(refreshToken)
			if err != nil || timeOut {
				log.Logger.Debug("valid refreshToken err:", zap.Error(err))
				//refreshToken出问题，表明用户三十天未登录，需要重新登录
				log.Logger.Info("user need login again")
				db.Redis.Del(context.Background(), token)
				c.Set("userId", "")
				c.Next()
				return
			}

			// refresh token 没过期
			userId, err := util.GetUserIDFormToken(refreshToken)
			if err != nil {
				log.Logger.Error("parse token to get uid error:", zap.Error(err))
				//token解析不了的情况一般很少,暂时panic一下
				panic(err)
			}

			//根据refreshToken 更新 accessToken
			accessToken, err := util.CreateAccessToken(userId)
			if err != nil {
				log.Logger.Error("create acc token error:", zap.Error(err))
				//token解析不了的情况一般很少,暂时panic一下
				panic(err)
			}

			//更新后，重新设置redis的key
			newRefreshToken, err := util.CreateRefreshToken(userId)
			if err != nil {
				log.Logger.Error("creat ref token error:", zap.Error(err))
				panic(err)
			}

			if err := db.Redis.Set(context.Background(), token, newRefreshToken, 30*24*time.Hour).Err(); err != nil {
				log.Logger.Error("create redis acc token error", zap.Error(err))
			} else {
				log.Logger.Debug("redis set success")
			}

			//后台登录更新token，本质上就是给login接口发送请求
			req := BackendLoginReq{}
			data, err := json.MarshalIndent(&req, "", "\t")
			if err != nil {
				log.Logger.Error("json parse error")
				c.Abort()
				return
			}
			request, err := http.NewRequest("POST", "http://localhost:8090/douyin/user/login?token="+accessToken, bytes.NewBuffer(data))
			if err != nil {
				log.Logger.Error("login move forward error")
				c.Abort()
				return
			}
			request.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			post, err := client.Do(request)
			if post.StatusCode == 200 {
				//发送登录请求成功
				c.Set("userId", userId)
				c.Next()
				return
			} else {
				log.Logger.Error("login move forward error")
				c.Abort()
				return
			}
		}
		//未过期
		userId, err := util.GetUserIDFormToken(token)
		if err != nil {
			panic(err)
		}
		c.Set("userId", userId)
		c.Next()
	}

}
