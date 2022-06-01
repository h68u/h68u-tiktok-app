package log

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"tikapp/common/config"
	"time"
)

// Logger 整个项目的Logger
var Logger *zap.Logger

func Init() {
	if config.AppCfg.RunMode == "debug" {
		// 开发模式 日志输出到终端
		core := zapcore.NewTee(
			zapcore.NewCore(getEncoder(),
				zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		)
		Logger = zap.New(core, zap.AddCaller())
	} else {
		fileLog()
	}
}

func fileLog() {
	// 调试级别
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	// 日志级别
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	// 警告级别
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	// 错误级别
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.ErrorLevel
	})
	// panic级别
	panicPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.PanicLevel
	})
	// fatal级别
	fatalPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.FatalLevel
	})

	cores := [...]zapcore.Core{
		getEncoderCore("./debug.log", debugPriority),
		getEncoderCore("./info.log", infoPriority),
		getEncoderCore("./warn.log", warnPriority),
		getEncoderCore("./error.log", errorPriority),
		getEncoderCore("./panic.log", panicPriority),
		getEncoderCore("./fatal.log", fatalPriority),
	}

	// zap.AddCaller() 可以获取到文件名和行号
	Logger = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller())
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func getLogWriter(fileName string) zapcore.WriteSyncer {
	dir, _ := os.Getwd() // 获取项目目录
	sperator0 := os.PathSeparator
	sperator := string(sperator0)
	// 	log 存放路径
	dir = dir + sperator + "runtime" + sperator + "logs"
	if !pathExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logrus.Warnf("create dir %s failed", dir)
		}
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   dir + sperator + fileName, // 日志文件路径
		MaxSize:    5,                         // 设置日志文件最大尺寸
		MaxBackups: 5,                         // 设置日志文件最多保存多少个备份
		MaxAge:     30,                        // 设置日志文件最多保存多少天
		Compress:   true,                      // 是否压缩 disabled by default
	}
	// 返回同步方式写入日志文件的zapcore.WriteSyncer
	return zapcore.AddSync(lumberJackLogger)
}

func getEncoderCore(fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	writer := getLogWriter(fileName)
	return zapcore.NewCore(getEncoder(), writer, level)
}

func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 将日志级别字符串转化为小写
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder, // 执行消耗时间转化成浮点型的秒
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 以包/文件:行号 格式化调用堆栈
	}
	return config
}

// CustomTimeEncoder 自定义日志输出时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(config.LogCfg.TimeFormat))
}
