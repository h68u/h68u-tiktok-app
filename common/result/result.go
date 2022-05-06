package result

import (
	"github.com/gin-gonic/gin"
	"reflect"
)

type Status struct {
	StatusCode int
	StatusMsg  string
}

func Success(data interface{}) interface{} {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice && value.IsNil() {
		data = gin.H{}
	}
	return &gin.H{
		"status_code": 0,
		"status_msg":  "success",
		"data":        data,
	}
}

func Error(status Status) interface{} {
	return gin.H{
		"status_code": status.StatusCode,
		"status_msg":  status.StatusMsg,
	}
}
