package pong

import (
	"gin_template/hub"
	"github.com/gin-gonic/gin"
)

func init() {
	instance = &pong{}
	hub.RegisterModule(instance)
}

var instance *pong

type pong struct {
}

func (m *pong) GetModuleInfo() hub.ModuleInfo {
	return hub.ModuleInfo{
		ID:       "internal.pong",
		Instance: instance,
	}
}

func (m *pong) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
}

func (m *pong) Serve(server *hub.Server) {
	// 注册服务函数部分
	server.HttpEngine.GET("/ping", handlePingPong)
}

func handlePingPong(c *gin.Context) {
	c.JSON(200, gin.H{
		"msg":        "pong",
		"User-Agent": c.GetHeader("User-Agent"),
	})
}
