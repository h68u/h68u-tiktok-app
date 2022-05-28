package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"

	"gorm.io/gorm"
)

type Favorite struct {
	UserId     int64  `form:"user_id"`
	Token      string `form:"token"`
	VideoId    int64  `form:"video_id"`
	ActionType int8   `form:"action_type"`
}

const (
	Unlike = iota
	Like
)

func DoFavorite(f *Favorite) error {
	if f.ActionType > 1 || f.ActionType < 0 {
		return errors.New("wrong action type")
	}

	tx := db.MySQL.Begin()

	// 检查是否存在点赞记录
	record := &model.VideoFavorite{}
	_ = tx.Model(&model.VideoFavorite{}).
		Where("user_id = ? and video_id = ?", f.UserId, f.VideoId).
		First(record)

	// 点赞
	if f.ActionType == Like {
		// 检查是否已经点赞
		if record.UserId != 0 {
			log.Logger.Error("repeat add favorite")
			return nil
		}

		// 插入点赞记录
		if err := tx.Create(&model.VideoFavorite{
			UserId:  f.UserId,
			VideoId: f.VideoId,
		}).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in creating video favorite record")
			return err
		}

		// 对应视频点赞数加一
		if err := tx.Model(&model.Video{}).
			Where("id = ?", f.VideoId).
			Update("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in adding video favorite count")
			return err
		}

		tx.Commit()
		return nil
	}

	// 取消点赞
	{
		// 检查是否存在点赞的记录
		if record.UserId == 0 {
			log.Logger.Error("repeat del favorite")
			return errors.New("")
		}

		// 删除点赞记录
		if err := tx.
			Where("user_id = ? and video_id = ?", f.UserId, f.VideoId).
			Delete(&model.VideoFavorite{}).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in deleting video favorite record")
			return err
		}

		// 对应视频点赞数减一
		if err := tx.Model(&model.Video{}).
			Where("id = ?", f.VideoId).
			Update("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in sub video favorite count")
			return err
		}
	}

	tx.Commit()
	return nil
}
