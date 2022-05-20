/*
	这里只是简单的封装 目的是方便调用
*/
package config

import (
	"fmt"
	"strings"
)

// GetInt32 用于获取配置中的整数类型
func GetInt32(key string) (int32, error) {
	rt := conf.GetInt32(key)
	if rt == 0 {
		return 0, fmt.Errorf("key required not found: %s", key)
	}
	return rt, nil
}

// GetString 获取字符串类型
func GetString(key string) (string, error) {
	rt := conf.GetString(key)
	if rt == "" {
		return "", fmt.Errorf("key required not found: %s", key)
	}
	return strings.ToLower(rt), nil
}
