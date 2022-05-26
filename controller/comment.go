package ctrl

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	res "tikapp/common/result"
	srv "tikapp/service"
)

type CommentReq struct {
	UserId      int64  `form:"user_id"`
	VideoId     int64  `form:"video_id"`
	ActionId    byte   `form:"action_type" `
	CommentId   int64  `form:"comment_id" `
	CommentText string `form:"comment_text" `
}

// CommentAction 执行评论
func CommentAction(c *gin.Context) {
	userId, _ := c.Get("UserId")
	fmt.Println(userId)
	var comm srv.Comment
	var req CommentReq
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		// todo log
		return
	}
	if req.ActionId != 1 && req.ActionId != 2 {
		res.Error(c, res.Status{
			StatusCode: res.QueryParamErrorStatus.StatusCode,
			StatusMsg:  res.QueryParamErrorStatus.StatusMsg,
		})
		return
	}

	// 发布评论
	if req.ActionId == 1 {
		comment, err := comm.Publish(req.UserId, req.VideoId, req.CommentText)
		if err != nil {
			// todo log
		}
		res.Success(c, res.R{
			"comment": comment,
		})
	}

	// 删除评论
	if req.ActionId == 2 {
		comment, err := comm.Delete(req.UserId, req.VideoId, req.CommentId)
		if err != nil {
			// todo log
			// 权限错误, 不允许删除其他用户评论
			if err == srv.ErrPermit {
				res.Error(c, res.Status{
					StatusCode: res.PermissionErrorStatus.StatusCode,
					StatusMsg:  res.PermissionErrorStatus.StatusMsg,
				})
			}

		}
		res.Success(c, res.R{
			"comment": comment,
		})
	}

}

// CommentList 查看视频所以评论 按发布时间倒序
func CommentList(c *gin.Context) {

}
