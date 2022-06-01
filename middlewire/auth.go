package middlewire

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"tikapp/common/db"
	"tikapp/common/log"
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
		validToken, err := util.ValidToken(token)

		if err != nil || validToken {
			//token过期或者解析token发生错误
			log.Logger.Info("token expire or parse token error")
			log.Logger.Info("valid refreshToken")
			value := db.Redis.Get(token)

			// 30d toke 能否取出
			refreshToken, err1 := value.Result()
			if err1 != nil {
				log.Logger.Error("token不合法，请确认你是否登录")
				c.JSON(200, gin.H{
					"status_code": 400,
					"msg":         "token不合法，请确认你操作是否有误",
				})
				c.Abort()
				return
			}

			// 可以取出30d token, 检查是否过期
			b, err1 := util.ValidToken(refreshToken)
			if err1 != nil || b {
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
			userId, err1 := util.GetUserIDFormToken(refreshToken)
			if err1 != nil {
				log.Logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}

			//根据refreshToken 更新 accessToken
			db.Redis.Del(token) // 首先删除redis记录
			accessToken, err1 := util.CreateAccessToken(userId)
			if err1 != nil {
				log.Logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}

			//更新后，重新设置redis的key
			newRefreshToken, err1 := util.CreateRefreshToken(userId)
			if err1 != nil {
				log.Logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}

			// 生成新的redis记录
			db.Redis.Set(accessToken, newRefreshToken, 30*24*time.Hour)

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
			url := c.Request.URL.String()
			split := strings.Split(url, "?") // eg.xxx:8080/api?a=1&b=2&token=xxx
			pre := split[0] + "?"            // eg. xxx:8080/api
			for key, val := range dataMap {
				pre = pre + key + "=" + val + "&"
			} // eg.xxx:8080/api?a=1&b=2&token=xxx&
			newUrl := strings.TrimSuffix(pre, "&") // eg.xxx:8080/api?a=1&b=2&token=xxx
			c.Redirect(http.StatusMovedPermanently, newUrl)
			c.Set("userId", userId)
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
