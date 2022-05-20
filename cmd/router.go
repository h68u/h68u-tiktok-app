package main

import "github.com/gin-gonic/gin"

func handle(r *gin.Engine) {

	basic := r.Group("/douyin")

	// 视频流
	feed := basic.Group("/feed")
	feed.GET("/", )

	// 用户相关
	userGroup := basic.Group("/user")
	{
		// 获取用户登录信息
		userGroup.GET("/", )

		// 新用户注册
		userGroup.POST("/register", )

		// 用户登录
		userGroup.POST("/register", )
	}

	// 视频投稿相关
	publishGroup := basic.Group("/publish")
	{
		publishGroup.GET("/publish", )
	}


	// 点赞相关
	favoriteGroup := basic.Group("/favorite")
	{
		favoriteGroup.POST("/", )
	}

	// 评论相关
	commentGroup := basic.Group("/comment")
	{
		commentGroup.GET("/", )
	}

	// 用户间关系操作 如关注 获取关注列表
	relationGroup := basic.Group("/relation")
	{
		relationGroup.GET("/", )
	}

}
