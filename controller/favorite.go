package ctrl

import (
	res "tikapp/common/result"
	srv "tikapp/service"

	"github.com/gin-gonic/gin"
)

// FavoriteAction 执行点赞和取消点赞操作
func FavoriteAction(c *gin.Context) {
	token, _ := c.Get("token")
	if token == "" {
		res.Error(c, res.Status{
			StatusCode: res.PermissionErrorStatus.StatusCode,
			StatusMsg:  res.PermissionErrorStatus.StatusMsg,
		})
		return
	}

	var req srv.Favorite
	err := c.ShouldBind(&req)
	if err != nil {
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if err = srv.DoFavorite(&req); err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FavoriteErrorStatus.StatusCode,
			StatusMsg:  res.FavoriteErrorStatus.StatusMsg,
		})
		return
	}

	res.Success(c, res.R{})
}

// FavoriteList 获取登录用户的所有点赞视频
func FavoriteList(c *gin.Context) {

	res.Success(c, res.R{})
}
