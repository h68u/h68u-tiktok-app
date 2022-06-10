package db

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var MySQL *gorm.DB
var Redis *redis.Client

func Init() {
	MySQLInit()
	RedisInit()
}