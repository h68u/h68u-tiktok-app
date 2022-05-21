/*
	model
*/
package model

import "tikapp/common/db"

type User struct {
	Id            int64  `gorm:"primaryKey"`
	Name          string `gorm:"index"`
	Username      string
	Password      string
	FollowCount   int64
	FollowerCount int64
	IsFollow      bool
}

func AutoCreateTable() {
	_ = db.DB.AutoMigrate(&User{})
}
