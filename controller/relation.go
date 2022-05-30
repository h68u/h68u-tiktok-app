package ctrl

import (
	"tikapp/common/log"
	res "tikapp/common/result"
	srv "tikapp/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var r srv.Relation

// RelationAction 关注或取消关注
func RelationAction(c *gin.Context) {
	var req srv.RelationFollow
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		log.Logger.Error("check params error")
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		log.Logger.Error("before login in")
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	if err = r.RelationAction(&req); err != nil {
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
	err := c.ShouldBindWith(&req, binding.Query)
	if err != nil {
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	var resp srv.UserFollowerResp
	if resp, err = r.FollowList(&req); err != nil {
		res.Error(c, res.Status{
			StatusCode: res.FollowListErrorStatus.StatusCode,
			StatusMsg:  res.FollowListErrorStatus.StatusMsg,
		})
		return
	}
	
	// resp := make(srv.UserFollowerResp, 0, len(resp0))
	// for i := 0; i < len(resp); i++ {
	// 	resp[i].Id = resp0[i].Id
	// 	resp[i].Name = resp0[i].Name
	// 	resp[i].FollowCount = resp0[i].FollowCount
	// 	resp[i].FollowerCount = resp0[i].FollowerCount
	// 	resp[i].IsFollow = true
	// }

	res.Success(c, res.R{
		"user_list": resp,
	})
}

// FollowerList 获取用户的粉丝列表
func FollowerList(c *gin.Context) {

}

