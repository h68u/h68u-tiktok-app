package ctrl

import (
	"tikapp/common/log"
	res "tikapp/common/result"
	srv "tikapp/service"

	"github.com/gin-gonic/gin"
)

var relationLogger = log.Namespace("RelationController")


// RelationAction 关注或取消关注
func RelationAction(c *gin.Context) {
	var r srv.Relation
	var req srv.RelationFollow
	err := c.ShouldBind(&req)
	if err != nil {
		relationLogger.Error("check params error")
		res.Error(c, res.QueryParamErrorStatus)
		return
	}

	if req.Token == "" {
		relationLogger.Error("before login in")
		res.Error(c, res.PermissionErrorStatus)
		return
	}

	if err = r.RelationAction(&req); err != nil {
		relationLogger.Error(err.Error())
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

}

// FollowerList 获取用户的粉丝列表
func FollowerList(c *gin.Context) {

}

