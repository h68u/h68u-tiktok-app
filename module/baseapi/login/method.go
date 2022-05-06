package login

import (
	"gin_template/common/model"
	"gin_template/common/result"
	"gin_template/common/statusCode"
	"gin_template/common/statusMsg"
	"gin_template/hub"
	"gin_template/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserLoginRequestBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

func login(c *gin.Context) {
	var userLoginRequestBody UserLoginRequestBody
	err := c.ShouldBindWith(&userLoginRequestBody, binding.JSON)
	if err != nil {
		result.Error(result.Status{
			StatusCode: statusCode.ParseJsonError,
			StatusMsg:  statusMsg.ParseJsonErrorMsg,
		})
	}
	var user model.User
	var count int64
	err = hub.DB.Model(&model.User{}).
		Where("username = ? and password = ?", userLoginRequestBody.Username, userLoginRequestBody.Password).
		Count(&count).First(&user).Error
	if err != nil {
		result.Error(result.Status{
			StatusCode: statusCode.MysqlError,
			StatusMsg:  statusMsg.MysqlErrorMsg,
		})
	}
	if count > 0 {
		var userLoginResponse UserLoginResponse
		userLoginResponse.UserId = user.Id
		token, err := utils.GenerateToken(user.Id)
		if err != nil {
			result.Error(result.Status{
				StatusCode: statusCode.JwtToken,
				StatusMsg:  statusMsg.JwtTokenMsg,
			})
		}
		userLoginResponse.Token = token
		result.Success(userLoginResponse)
	}
}
