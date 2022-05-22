package model

type Video struct {
	Id int64 `gorm:"primaryKey"`
	PublishId int64 
	PlayUrl string
	CoverUrl string
	// CreateTime 
}