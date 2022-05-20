package config

import (
	"github.com/sirupsen/logrus"
	"time"
)

type App struct {
	Host      string
	Port      string
	JwtSecret string
	Release   string
	RunMode   string
}

type Mysql struct {
	Address     string
	User        string
	Password    string
	DBName      string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
}

type Redis struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxOpen     int
	IdleTimeOut time.Duration
}

type Log struct {
	LogSavePath string
	LogFileExt  string
	TimeFormat  string
}

var (
	AppCfg   App
	MysqlCfg Mysql
	RedisCfg Redis
	LogCfg   Log
)

func Init() {
	var err error

	if err = conf.UnmarshalKey("app", &AppCfg); err != nil {
		logrus.Panicf("parse config err, app: %v", err)
	}

	if err = conf.UnmarshalKey("db.mysql", &MysqlCfg); err != nil {
		logrus.Panicf("parse config err, mysql: %v", err)
	}

	if err = conf.UnmarshalKey("db.redis", &RedisCfg); err != nil {
		logrus.Panicf("parse config err, redis: %v", err)
	}

	if err = conf.UnmarshalKey("log", &LogCfg); err != nil {
		logrus.Panicf("parse config err, log: %v", err)
	}

	logrus.Debug("parse config success")
}
