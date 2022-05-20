package util

import (
	"fmt"
	"tikapp/common/config"
)

// GetServerLoc 用于获取配置文件 app.yaml 中 server 的配置
// 将地址和端口拼接后返回给 gin
func GetServerLoc() string {
	addr, err := config.GetString("server.addr")
	if err != nil {
		panic(err)
	}
	port, err := config.GetInt32("server.port")
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s:%d", addr, port)
}