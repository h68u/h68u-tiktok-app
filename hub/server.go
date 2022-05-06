package hub

import (
	"gin_template/common/constant"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Instance *Server

type Server struct {
	HttpEngine *gin.Engine
}

var logger = logrus.WithField("server", "internal")

// Init 快速初始化
func Init() {
	gin.SetMode(gin.ReleaseMode)
	httpEngine := gin.New()
	httpEngine.Use(ginRequestLog(), gin.Recovery())
	Instance = &Server{
		HttpEngine: httpEngine,
	}
	InitDatabase()
}

// StartService 启动服务
// 根据 Module 生命周期 此过程应在Login前调用
// 请勿重复调用
func StartService() {
	logger.Infof("initializing modules ...")
	for _, mi := range modules {
		mi.Instance.Init()
	}
	logger.Info("all modules initialized")

	logger.Info("registering modules serve functions ...")
	for _, mi := range modules {
		mi.Instance.Serve(Instance)
	}
	logger.Info("all modules serve functions registered")
}

// Run 正式开启服务
func Run() {
	go func() {
		logger.Info("http engine starting...")
		if err := Instance.HttpEngine.Run(constant.ServerPort); err != nil {
			logger.Fatal(err)
		} else {
			logger.Info("http engine running...")
		}
	}()
}
