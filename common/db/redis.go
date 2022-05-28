package db

import (
	"tikapp/common/config"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func RedisInit() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.RedisCfg.Host,
		Password: config.RedisCfg.Password,
		DB:       0,
		IdleTimeout: config.RedisCfg.IdleTimeOut,
	})
	if _, err := Redis.Ping().Result(); err != nil {
		logrus.Panic("connect redis failed: %v", err)
	}
	logrus.Info("Connect redis succeeded")
}
//更新redis
func UpdateRedis(){

}

//更新DB
func UpdateMysql{

}