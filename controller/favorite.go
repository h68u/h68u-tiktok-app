package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	srv "tikapp/service"
	res "tikapp/common/result"
)

type FavoriteActionReq struct{
	UserId		int64		`form:"user_id"`
	Token       string		`form:"token"`
	VideoId		int64		`form:"video_id"`
	ActionId	byte		`form:"action_type"`
}

type FavoriteListReq struct{
	UserId		int64		`form:"user_id"`
	Token 		token       `form:"token"`
}

var favorite srv.VideoFavorite
// FavoriteAction 执行点赞和取消点赞操作
func FavoriteAction(c *gin.Context) {
	var req FavoriteActionReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		log.Logger.Error("parse json error")
		res.Error(c, res.Status{
			StatusCode: res.ServerErrorStatus.StatusCode,
			StatusMsg:  res.ServerErrorStatus.StatusMsg,
		})
		return
	}
	//鉴权？
	// 请求参数错误
	if req.ActionId != 0 && req.ActionId != 1 {
		res.Error(c, res.Status{
			StatusCode: res.QueryParamErrorStatus.StatusCode,
			StatusMsg:  res.QueryParamErrorStatus.StatusMsg,
		})
		return
	}

	switch req.ActionId{
	case 1:
		//点赞
		resp, err := favorite.SetFavor(req.VideoId,req.UserId)
		if err != nil{
			return 
		}
	case 0:
		//取消赞
		resp,err = favorite.RemoveFavor(req.VideoId,req.UserId)
		if err != nil{
			return
		}
		
	}

	res.Success1(c)

	

}

// FavoriteList 获取登录用户的所有点赞视频
func FavoriteList(c *gin.Context) {
	var req FavoriteListReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		log.Logger.Error("parse json error")
		res.Error(c, res.Status{
			StatusCode: res.ServerErrorStatus.StatusCode,
			StatusMsg:  res.ServerErrorStatus.StatusMsg,
		})
		return
	}
	favorlist, err := comm.List(req.UserId)
	res.Success(c, res.R{
		"VideoDemo": favorlist,
	})
}
