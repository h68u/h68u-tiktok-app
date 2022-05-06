package hub

import (
	"gin_template/common/config"
	"gin_template/common/constant"
	"gin_template/common/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/**
数据库初始化的地方
*/

var (
	DB *gorm.DB
)

func InitDatabase() {
	InitMysql()
}

func InitMysql() {
	username := config.Conf.GetString(constant.MysqlUsername)
	password := config.Conf.GetString(constant.MysqlPassword)
	addr := config.Conf.GetString(constant.MysqlAddr)
	dbname := config.Conf.GetString(constant.MysqlDbName)
	db, err := gorm.Open(mysql.Open(username + ":" + password + "@tcp(" + addr + ")/" + dbname + "?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		logger.Error("connect mysql error:" + err.Error())
		panic("application start fail")
	}
	logger.Info("mysql connect success")
	DB = db
	/**
	在此添加实体类，也可以抽取出来
	*/
	err = db.AutoMigrate(
		&model.User{})
	if err != nil {
		panic(err)
	}
	logger.Info("mysql init database success")
}
