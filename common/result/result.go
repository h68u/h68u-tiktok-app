package res

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	SuccessStatus           = newStatus(200, "success")
	LoginErrorStatus        = newStatus(400, "login happen error")
	RegisterErrorStatus     = newStatus(401, "register happen error")
	UsernameExitErrorStatus = newStatus(402, "username already exists")
	TokenErrorStatus        = newStatus(403, "token error")
	InfoErrorStatus         = newStatus(404, "can't get user info")
	NoLoginErrorStatus      = newStatus(405, "user no login")
	FileErrorStatus         = newStatus(406, "upload file error")
	PublishErrorStatus      = newStatus(407, "publish error")
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

func Error(c *gin.Context, status Status) {
	c.JSON(http.StatusOK, gin.H{
		"status_code": status.StatusCode,
		"status_msg":  status.StatusMsg,
	})
}
