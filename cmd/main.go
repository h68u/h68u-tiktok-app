package main

import (
	"fmt"
	"sync"
	"tikapp/common/config"
	"tikapp/common/cron"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/oss"

	"github.com/gin-gonic/gin"
)

var once sync.Once

func init() {
	once.Do(func() {
		config.ReadCfg()
		config.Init()
		log.Init()
		db.Init()
		oss.Init()
		cron.Init()
	})
}

func main() {
	r := gin.Default()

	handle(r)

	r.Run(fmt.Sprintf("%s:%s", config.AppCfg.Host, config.AppCfg.Port))
}
