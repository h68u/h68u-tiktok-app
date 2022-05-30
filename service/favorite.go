package srv

import (
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
)

// IsFavorite 判断某个用户是否收藏了某个视频，为 service\publish.go 中的 PublishList 方法提供依赖
func IsFavorite(userId, videoId int64) (bool, error) {
	var count int64
	err := db.MySQL.Debug().Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", userId, videoId).Count(&count).Error
	if err != nil {
		log.Logger.Error("mysql happen error when check favorite")
		return false, err
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}
