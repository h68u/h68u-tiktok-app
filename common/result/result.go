package res

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	SuccessStatus           = newStatus(200, "success")
	LoginErrorStatus        = newStatus(400, "登录发生错误")
	RegisterErrorStatus     = newStatus(401, "注册发生错误")
	UsernameExitErrorStatus = newStatus(402, "用户名已存在")
	TokenErrorStatus        = newStatus(403, "token 错误")
	InfoErrorStatus         = newStatus(404, "无法获取该用户信息")
	NoLoginErrorStatus      = newStatus(405, "用户未登录")
	FileErrorStatus         = newStatus(406, "文件上传失败")
	PublishErrorStatus      = newStatus(407, "发布时出现错误")

	FeedErrorStatus = newStatus(409, "获取视频流出错")

	EmptyErrorStatus = newStatus(408, "用户名或密码为空")

	QueryParamErrorStatus     = newStatus(409, "请求的参数错误")
	PermissionErrorStatus     = newStatus(410, "permission error")
	CommentNotExitErrorStatus = newStatus(411, "评论不存在")
	VideoNotExitErrorStatus   = newStatus(412, "视频不存在")
	FollowErrorStatus         = newStatus(413, "关注失败")
	FavoriteErrorStatus       = newStatus(414, "点赞失败")
	FollowListErrorStatus     = newStatus(415, "获取关注列表时发生了错误")

	ServerErrorStatus = newStatus(500, "服务器内部错误")
)

type Status struct {
	StatusCode int64
	StatusMsg  string
}

func (s Status) Code() int64 {
	return s.StatusCode
}

func (s Status) Mag() string {
	return s.StatusMsg
}

func newStatus(code int64, msg string) Status {
	return Status{code, msg}
}

type R map[string]interface{}

func Success(c *gin.Context, r R) {
	//value := reflect.ValueOf(data)
	h := gin.H{
		"status_code": 0,
		"status_msg":  "success",
	}
	for s, v := range r {
		h[s] = v
	}
	c.JSON(http.StatusOK, h)
}
func Success1(c *gin.Context) {
	h := gin.H{
		"status_code": 0,
		"status_msg":  "success",
	}
	c.JSON(http.StatusOK, h)
}

func Error(c *gin.Context, status Status) {
	c.JSON(http.StatusOK, gin.H{
		"status_code": status.StatusCode,
		"status_msg":  status.StatusMsg,
	})
}
