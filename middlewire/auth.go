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

var logger = log.NameSpace("auth")

//鉴权接口
func Auth() gin.HandlerFunc {
	//先判断请求头是否为空，为空则为游客状态
	return func(c *gin.Context) {
		method := c.Request.Method
		token := ""
		if method == "GET" {
			token = c.DefaultQuery("token", "")
		} else {
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
		//有token，判断是否过期
		validToken, err := util.ValidToken(token)
		if err != nil || validToken {
			//token过期或者解析token发生错误
			logger.Info("token expire or parse token error")
			logger.Info("valid refreshToken")
			value := db.Redis.Get(token)
			refreshToken, err1 := value.Result()
			if err1 != nil {
				logger.Error("token不合法，请确认你是否登录")
				c.JSON(200, gin.H{
					"status_code": 400,
					"msg":         "token不合法，请确认你操作是否有误",
				})
				c.Abort()
				return
			}
			b, err1 := util.ValidToken(refreshToken)
			if err1 != nil || b {
				//refreshToken出问题，表明用户三十天未登录，需要重新登录
				logger.Info("user need login again")
				//直接变成访客状态
				/*u := uuid.New()
				c.SetCookie("visit-user", u.String(), 30*24*60*60*1000, "/", "localhost", false, true)
				c.Next()*/
				db.Redis.Del(token)
				c.Set("userId", "")
				c.Next()
				return
			}
			//根据refreshToken刷新accessToken
			userId, err1 := util.GetUsernameFormToken(refreshToken)
			if err1 != nil {
				logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}
			//刷新后，重新设置redis的key
			db.Redis.Del(token)
			accessToken, err1 := util.CreateAccessToken(userId)
			if err1 != nil {
				logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}
			newRefreshToken, err1 := util.CreateRefreshToken(userId)
			if err1 != nil {
				logger.Error("parse token error")
				//token解析不了的情况一般很少,暂时panic一下
				panic(err1)
			}
			db.Redis.Set(accessToken, newRefreshToken, 30*24*time.Hour)
			//获取之前请求的所有query参数
			dataMap := make(map[string]string)
			for key, _ := range c.Request.URL.Query() {
				if key == "token" {
					dataMap[key] = accessToken
				} else {
					dataMap[key] = c.Query(key)
				}
			}
			//转发路由携带新token
			url := c.Request.URL.String()
			split := strings.Split(url, "?")
			pre := split[0]
			for key, val := range dataMap {
				pre = pre + "?" + key + "=" + val + "&"
			}
			newUrl := strings.TrimSuffix(pre, "&")
			c.Redirect(http.StatusMovedPermanently, newUrl)
			c.Set("userId", userId)
			c.Next()
			return
		}
		//未过期
		userId, err := util.GetUsernameFormToken(token)
		if err != nil {
			panic(err)
		}
		c.Set("userId", userId)
		c.Next()
		return
	}

}
