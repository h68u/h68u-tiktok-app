package main

import (
	ctrl "tikapp/controller"
	"tikapp/middlewire"

	"github.com/gin-gonic/gin"
)

func handle(r *gin.Engine) {
	// TODO: 或许可以在这加一个全局处理 err 的 handler
	// r.Any("/", ctrl.ErrorHanler)

	r.Any("/log", ctrl.LogToWeb)

	// 鉴权
	r.Use(middlewire.Auth())

	// 测试
	r.Any("/ping", ctrl.Ping)

	basic := r.Group("/douyin")

	// 视频流
	feed := basic.Group("/feed", ctrl.Feed)
	feed.GET("/")

	// 用户相关
	userGroup := basic.Group("/user")
	{
		// 获取用户登录信息
		userGroup.GET("/", ctrl.Info)

		// 新用户注册
		userGroup.POST("/register/", ctrl.Register)

		// 用户登录
		userGroup.POST("/login/", ctrl.Login)
	}

	// 视频投稿相关
	publishGroup := basic.Group("/publish")
	{
		// 用户上传视频
		publishGroup.POST("/action/", ctrl.PublishAction)

		// 直接列出用户投稿过的所有视频
		publishGroup.GET("/list/", ctrl.PublishList)
	}

	// 点赞相关
	favoriteGroup := basic.Group("/favorite")
	{
		// 点赞 取消点赞
		favoriteGroup.POST("/action/", ctrl.FavoriteAction)

		// 获取点赞列表
		favoriteGroup.GET("/list/", ctrl.FavoriteList)
	}

	// 评论相关
	commentGroup := basic.Group("/comment")
	{
		// 评论
		commentGroup.POST("/action/", ctrl.CommentAction)

		// 倒叙查看评论
		commentGroup.GET("/list/", ctrl.CommentList)
	}

	// 用户间关系操作 如关注 获取关注列表
	relationGroup := basic.Group("/relation")
	{
		// 对指定用户 关注 取关
		relationGroup.POST("/action/", ctrl.RelationAction)

		// 获取用户的关注列表
		relationGroup.GET("/follow/list/", ctrl.FollowList)

		// 获取用户的粉丝列表
		relationGroup.GET("/follower/list/", ctrl.FollowerList)
	}

}
