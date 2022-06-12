package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"tikapp/common/log"
	res "tikapp/common/result"
	srv "tikapp/service"
)

type CommentActionReq struct {
	VideoId     int64  `form:"video_id" binding:"required"`
	ActionId    byte   `form:"action_type" binding:"required"`
	CommentId   int64  `form:"comment_id"`
	CommentText string `form:"comment_text"`
	Token       string `form:"token" binding:"required"`
}

type CommentListReq struct {
	VideoId int64  `form:"video_id" binding:"required"`
	Token   string `form:"token" `
}

var comm srv.Comment

// CommentAction 执行评论
// todo 错误处理有点繁琐, 之后加个中间件处理
func CommentAction(c *gin.Context) {
	userIdI, _ := c.Get("userId")
	userId := userIdI.(int64)
	var req CommentActionReq
	err := c.ShouldBindWith(&req, binding.Query)
	if req.Token == "" {
		log.Logger.Error("operation illegal")
		res.Error(c, res.PermissionErrorStatus)
		return
	}
	if err != nil {
		log.Logger.Error("parse json error")
		res.Error(c, res.Status{
			StatusCode: res.ServerErrorStatus.StatusCode,
			StatusMsg:  res.ServerErrorStatus.StatusMsg,
		})
		return
	}
	if req.ActionId != 1 && req.ActionId != 2 {
		log.Logger.Error("wrong action type")
		res.Error(c, res.Status{
			StatusCode: res.QueryParamErrorStatus.StatusCode,
			StatusMsg:  res.QueryParamErrorStatus.StatusMsg,
		})
		return
	}

	var commentResp srv.CommentResp

	switch req.ActionId {
	// 发布评论
	case 1:
		commentResp, err = comm.Publish(userId, req.VideoId, req.CommentText)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				res.Error(c, res.Status{
					StatusCode: res.VideoNotExitErrorStatus.StatusCode,
					StatusMsg:  res.VideoNotExitErrorStatus.StatusMsg,
				})
			}

			return
		}
	// 删除评论
	case 2:
		commentResp, err = comm.Delete(userId, req.VideoId, req.CommentId)
		if err != nil {
			// 评论不存在
			if err == gorm.ErrRecordNotFound {
				res.Error(c, res.Status{
					StatusCode: res.CommentNotExitErrorStatus.StatusCode,
					StatusMsg:  res.CommentNotExitErrorStatus.StatusMsg,
				})
				return
			}
			// 权限错误, 不允许删除其他用户评论
			if err == srv.ErrPermit {
				log.Logger.Error(err.Error())
				res.Error(c, res.Status{
					StatusCode: res.PermissionErrorStatus.StatusCode,
					StatusMsg:  res.PermissionErrorStatus.StatusMsg,
				})
				return
			}
		}
	}

	res.Success(c, res.R{
		"comment": commentResp,
	})

}

// CommentList 查看视频所以评论 按发布时间倒序
func CommentList(c *gin.Context) {
	var req CommentListReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		log.Logger.Error("parse json error")
		res.Error(c, res.Status{
			StatusCode: res.ServerErrorStatus.StatusCode,
			StatusMsg:  res.ServerErrorStatus.StatusMsg,
		})
		return
	}
	comments, err := comm.List(req.VideoId)
	res.Success(c, res.R{
		"comment_list": comments,
	})
}
