package db

import (
	"tikapp/common/config"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func RedisInit() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.RedisCfg.Host,
		Password: config.RedisCfg.Password,
		DB:       0,
		IdleTimeout: -1,
	})
	if _, err := Redis.Ping().Result(); err != nil {
		logrus.Panic("connect redis failed: %v", err)
	}
	logrus.Info("Connect redis succeeded")
	
	// 防止 client 挂掉 应该有更优雅的方法，现在这样勉强能用 (大概)
	go func() {
		for {
			time.Sleep(time.Minute * 90)
			Redis.Ping()
		}
		// for {
		// 	select {
		// 	case <-time.After(time.Minute * 90):
		// 		Redis.Ping()
		// 	}
		// }
	}()
}
//更新redis
func UpdateRedis(){

}

//更新DB
func UpdateMysql(){

}