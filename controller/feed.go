package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	res "tikapp/common/result"
	srv "tikapp/service"
)

// Feed 不限制登录状态 按投稿时间获取视频流 单次最多返回 30 个
func Feed(c *gin.Context) {
	userId, _ := c.Get("userId")
	var id int64
	if userId == "" {
		id = -1
	} else {
		id = userId.(int64)
	}

	var lastTime int64
	if c.Query("lastTime") != "" {
		lastTime, _ = com.StrTo(c.Query("lastTime")).Int64()
	}
	var f srv.Feed
	resp, err := f.Feed(id, lastTime)
	if err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FeedErrorStatus.StatusCode,
			StatusMsg:  res.FeedErrorStatus.StatusMsg,
		})
		return
	}
	feedResp := resp.(srv.FeedResp)
	res.Success(c, res.R{
		"next_time":  feedResp.NextTime,
		"video_list": feedResp.VideoList,
	})
}
