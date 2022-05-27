package log

import (
	"unsafe"

	"go.uber.org/zap"
)

type namespace string

func (n namespace) Error(msg string, args ...zap.Field) {
	if args != nil {
		a := make([]zap.Field, 0, len(args)+1)
		a = append(a, zap.Namespace(*(*string)(unsafe.Pointer(&n))))
		a = append(a, args...)
		logger.Error(msg, a...)
	}
	logger.Error(msg)
}

func (n namespace) Info(msg string, args ...zap.Field) {
	if args != nil {
		a := make([]zap.Field, 0, len(args)+1)
		a = append(a, zap.Namespace((*(*string)(unsafe.Pointer(&n)))))
		a = append(a, args...)
		logger.Info(msg, a...)
	}
	logger.Info(msg, args...)
}

func (n namespace) Warn(msg string, args ...zap.Field) {
	if args != nil {
		a := make([]zap.Field, 0, len(args)+1)
		a = append(a, zap.Namespace(*(*string)(unsafe.Pointer(&n))))
		a = append(a, args...)
		logger.Warn(msg, a...)
	}
	logger.Warn(msg)
}

func (n *namespace) Debug(msg string, args ...zap.Field) {
	if args != nil {
		a := make([]zap.Field, 0, len(args)+1)
		a = append(a, zap.Namespace(*(*string)(unsafe.Pointer(&n))))
		a = append(a, args...)
		logger.Debug(msg, a...)
	}
	logger.Debug(msg, args...)
}

// Namespace 返回一个带有名称的执行全局 logger 的对象
// 用于给定 Namespace 的日志记录
// 日志将直接输出到日志文件
func Namespace(n string) namespace {
	return namespace(n)
}
