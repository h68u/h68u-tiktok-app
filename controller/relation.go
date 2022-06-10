package ctrl

import (
	"tikapp/common/log"
	res "tikapp/common/result"
	srv "tikapp/service"

	"github.com/gin-gonic/gin"
)

var r srv.Relation

// RelationFollowReq 关注请求
type RelationFollowReq struct {
	Token      string `form:"token" binding:"required"`
	ToUserId   int64  `form:"to_user_id" binding:"required"`
	ActionType int32  `form:"action_type" binding:"required"`
}

// RelationAction 关注或取消关注
func RelationAction(c *gin.Context) {
	var req RelationFollowReq
	var srvR srv.RelationFollow
	err := c.ShouldBind(&req)
	if err != nil {
		log.Logger.Error("check params error")
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		log.Logger.Error("operation illegal")
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	userId, _ := c.Get("userId")
	srvR.UserId = userId.(int64)
	srvR.ToUserId = req.ToUserId
	srvR.Token = req.Token
	srvR.ActionType = req.ActionType

	if err = r.RelationAction(&srvR); err != nil {
		log.Logger.Error(err.Error())
		res.Error(c, res.Status{
			StatusCode: res.FollowErrorStatus.StatusCode,
			StatusMsg:  res.FollowErrorStatus.StatusMsg,
		})
		return
	}

	res.Success(c, res.R{})
}

// FollowList 获取用户关注的列表
func FollowList(c *gin.Context) {
	var req srv.UserFollowerReq
	err := c.ShouldBind(&req)
	if err != nil {
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		log.Logger.Error("operation illegal")
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	t, _ := c.Get("userId")
	userId := t.(int64)

	var resp srv.UserFollowerResp
	if resp, err = r.FollowList(&req, userId); err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FollowListErrorStatus.StatusCode,
			StatusMsg:  res.FollowListErrorStatus.StatusMsg,
		})
		return
	}

	res.Success(c, res.R{
		"user_list": resp,
	})
}

// FollowerList 获取用户的粉丝列表
func FollowerList(c *gin.Context) {
	var req srv.UserFollowerReq
	err := c.ShouldBind(&req)
	if err != nil {
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		log.Logger.Error("operation illegal")
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	ans, err := srv.FollowerList(req.UserId)
	if err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FollowListErrorStatus.StatusCode,
			StatusMsg:  res.FollowListErrorStatus.StatusMsg,
		})
		return
	}

	res.Success(c, res.R{
		"user_list": ans,
	})
}
