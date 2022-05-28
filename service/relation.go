package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"

	"gorm.io/gorm"
)

var relationLogger = log.Namespace("RelationService")

type Relation struct{}

type RelationFollow struct {
	UserId     int64  `form:"user_id" binding:"required"`
	Token      string `form:"token" binding:"required"`
	ToUserId   int64  `form:"to_user_id" binding:"required"`
	ActionType int8   `form:"action_type" binding:"required"`
}

const (
	doFollow = iota + 1
	unFollow
)

func (r Relation) RelationAction(d *RelationFollow) error {
	if d.UserId == d.ToUserId {
		relationLogger.Error("self operation")
		return errors.New("self operation is not allowed")
	}

	if d.ActionType > 2 || d.ActionType < 1 {
		return errors.New("wrong action type")
	}

	var rel model.Follow
	db.MySQL.Debug().
		Model(&model.Follow{}).
		Where("follow_id = ? and user_id = ?", d.UserId, d.ToUserId).
		First(&rel)

	tx := db.MySQL.Begin()
	rel.FollowId = d.UserId
	rel.UserId = d.ToUserId

	// 关注
	if d.ActionType == doFollow && rel.CreateTime == 0 {
		// 加入关注列表
		if err := tx.Debug().Create(&rel).Error; err != nil {
			relationLogger.Error("mysql error in doing follow action")
			return err
		}
		
		// 更新关注者关注的人数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.FollowId).
			Update("follow_count", gorm.Expr("follow_count + ?", 1)).Error;
		err != nil {
			tx.Rollback()
			relationLogger.Error("mysql error in updating follow_count")
			return err
		}

		// 更新被关注者被关注数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.UserId).
			Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error;
		err != nil {
			tx.Rollback()
			relationLogger.Error("mysql error in updating follower_count")
			return err
		}
	}

	// 取关
	if d.ActionType == unFollow && rel.CreateTime != 0 {
		// 删除关注记录
		if err := tx.Debug().Delete(&rel).Error; err != nil {
			tx.Rollback()
			relationLogger.Error("mysql error in doing unfollow action")
			return err
		}

		// 更新关注者关注的人数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.FollowId).
			Update("follow_count", gorm.Expr("follow_count - ?", 1)).Error;
		err != nil {
			tx.Rollback()
			relationLogger.Error("mysql error in updating follow_count")
			return err
		}

		// 更新被关注者被关注数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.UserId).
			Update("follower_count", gorm.Expr("follower_count - ?", 1)).Error;
		err != nil {
			tx.Rollback()
			relationLogger.Error("mysql error in updating follower_count")
			return err
		}
	}

	tx.Commit()
	return nil
}
