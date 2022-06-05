package middlewire

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"tikapp/common/db"
	"tikapp/common/log"
	srv "tikapp/service"
	"tikapp/util"
	"time"
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
// TODO 目前redis更新时可能并发不安全（能力不够，不知道怎么解决）
func Auth() gin.HandlerFunc {
	//先判断请求头是否为空，为空则为游客状态
	return func(c *gin.Context) {
		token := ""
		token = c.DefaultQuery("token", "")
		if token == "" {
			token = c.PostForm("token")
		}
		if token == "" {

			/*//判断浏览器是否存在cookie，存在表示非第一次访问
			logger.Info("start valid cookie")
			_, err := c.Cookie("visit-user")
			if err != nil {
				//没有这个cookie，第一次访问
				logger.Info("first visit")
				//生成唯一id，作为游客的userId
				u := uuid.New()
				c.Set("userId","")
				//TODO cookie有效期待定
				c.SetCookie("visit-user", u.String(), 30*24*60*60*1000, "/", "localhost", false, true)
				c.Next()
				return
			}
			//有cookie,直接next*/

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
			value := db.Redis.Get(token)
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

				//直接变成访客状态
				/*u := uuid.New()
				c.SetCookie("visit-user", u.String(), 30*24*60*60*1000, "/", "localhost", false, true)
				c.Next()*/

				db.Redis.Del(token)
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

			/*
				是否应该删除旧的token
				如果一直使用旧的token请求，那么一个折中的方法是更新kv
				（old acc, old ref）更新为 （old acc, new ref）

				目前解决方法：
				删除旧的token, 调用登录接口后台帮助 非登录30天的用户登录
			*/
			// TODO  新的解决方法
			//db.Redis.Del(token)

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
				//token解析不了的情况一般很少,暂时panic一下
				panic(err)
			}

			// debug
			//{
			//	log.Logger.Debug("old acc token: " + token)
			//	log.Logger.Debug("new acc token: " + accessToken)
			//	log.Logger.Debug("new ref token: " + newRefreshToken)
			//}

			if err := db.Redis.Set(token, newRefreshToken, 30*24*time.Hour).Err(); err != nil {
				log.Logger.Error("create redis acc token error", zap.Error(err))
			} else {
				log.Logger.Debug("redis set success")
			}

			// 生成新的redis记录
			// TODO 新的解决方法: 后台登录
			//if err := db.Redis.Set(token, newRefreshToken, 30*24*time.Hour).Err(); err != nil {
			//	log.Logger.Error("create redis acc token error", zap.Error(err))
			//} else {
			//	log.Logger.Debug("redis set success")
			//}

			//获取之前请求的所有query参数： 替换过期的acc(ref还未失效期间)，过期的acc仍然可以使用接口
			dataMap := make(map[string]string)
			for key := range c.Request.URL.Query() {
				if key == "token" { // 修改之前所有的token
					dataMap[key] = accessToken
				} else {
					dataMap[key] = c.Query(key)
				}
			}

			//转发路由携带新token
			url1 := c.Request.URL.String()
			split := strings.Split(url1, "?") // eg.xxx:8080/api?a=1&b=2&token=xxx
			pre := split[0] + "?"             // eg. xxx:8080/api
			for key, val := range dataMap {
				pre = pre + key + "=" + val + "&"
			} // eg.xxx:8080/api?a=1&b=2&token=xxx&
			newUrl := strings.TrimSuffix(pre, "&") // eg.xxx:8080/api?a=1&b=2&token=xxx

			log.Logger.Debug("check url", zap.String("newUrl", newUrl))

			c.Redirect(http.StatusMovedPermanently, newUrl)
			c.Set("userId", userId)

			// TODO 后台登录
			req := srv.UserLoginReq{
				Username: "",
				Password: "",
				Token:    accessToken,
			}
			data, err := json.MarshalIndent(&req, "", "\t")
			if err != nil {
				log.Logger.Error("json parse error")
				c.Abort()
				return
			}
			request, err := http.NewRequest("POST", "http://localhost:8090/douyin/user/login?", bytes.NewBuffer(data))
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
			// 要求下次请求更换url
			//log.Logger.Debug("backend login start")
			//ctrl.Login(c)
			//log.Logger.Debug("backend login finish")
			c.Next()
			return
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
