package model

type User struct {
	Id            int64 `gorm:"primaryKey"`
	Name          string
	Username      string
	Password      string
	FollowCount   int64
	FollowerCount int64
}
