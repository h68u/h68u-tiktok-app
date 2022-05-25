package ctrl

import (
	"github.com/gin-gonic/gin"
	res "tikapp/common/result"
	srv "tikapp/service"
)

// PublishAction 已登录的用户上传视频
func PublishAction(c *gin.Context) {
	userId, _ := c.Get("userId")
	if userId == "" {
		res.Error(c, res.Status{
			StatusCode: res.NoLoginErrorStatus.StatusCode,
			StatusMsg:  res.NoLoginErrorStatus.StatusMsg,
		})
		return
	}
	title := c.PostForm("title")
	data, err := c.FormFile("data")

	if err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FileErrorStatus.StatusCode,
			StatusMsg:  res.FileErrorStatus.StatusMsg,
		})
		return
	}
	var v srv.Video
	err = v.PublishAction(data, title, userId.(int64))
	if err != nil {
		res.Error(c, res.Status{
			StatusCode: res.PublishErrorStatus.StatusCode,
			StatusMsg:  res.PublishErrorStatus.StatusMsg,
		})
		return
	}
	res.Success(c, res.R{})
}

// PublishList 列出当前用户所有的投稿视频
func PublishList(c *gin.Context) {

}
