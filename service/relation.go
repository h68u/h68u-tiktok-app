package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"

	"gorm.io/gorm"
)

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
		log.Logger.Error("self operation")
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
			log.Logger.Error("mysql error in doing follow action")
			return err
		}

		// 更新关注者关注的人数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.FollowId).
			Update("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in updating follow_count")
			return err
		}

		// 更新被关注者被关注数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.UserId).
			Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in updating follower_count")
			return err
		}
	}

	// 取关
	if d.ActionType == unFollow && rel.CreateTime != 0 {
		// 删除关注记录
		if err := tx.Debug().Delete(&rel).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in doing unfollow action")
			return err
		}

		// 更新关注者关注的人数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.FollowId).
			Update("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in updating follow_count")
			return err
		}

		// 更新被关注者被关注数
		if err := tx.Debug().Model(&model.User{}).
			Where("id = ?", rel.UserId).
			Update("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			log.Logger.Error("mysql error in updating follower_count")
			return err
		}
	}

	tx.Commit()
	return nil
}

type UserFollowerReq struct {
	UserId int64  `form:"user_id"`
	Token  string `form:"token"`
}

type UserFollowerItem0 struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
}

type UserFollowerItem struct {
	UserFollowerItem0

	IsFollow bool `json:"is_follow"`
}

type UserFollowerResp0 = []UserFollowerItem0
type UserFollowerResp = []UserFollowerItem

const sqlUserFollower = `select u.id, u.name, u.follow_count, u.follower_count from user as u, follow as f where f.follow_id = ? and f.user_id = u.id;`

// FollowList 获取给定用户的关注列表
func FollowList(u *UserFollowerReq) (UserFollowerResp, error) {
	var uResp UserFollowerResp

	rows, err := db.MySQL.Debug().Raw(sqlUserFollower, u.UserId).Rows()
	if err != nil {
		log.Logger.Error("mysql error in get follower list")
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	// for i := 0; rows.Next(); i++ {
	// 	if err := rows.Scan(&uResp0[i].Id, &uResp0[i].Name, &uResp[i].FollowCount, &uResp[i].FollowerCount);
	// 	err != nil {
	// 		log.Logger.Error("mysql error in writing in uResp0")
	//      return nil, err
	// 	}
	// }

	for rows.Next() {
		bucket := struct {
			Id            int64
			Name          string
			FollowCount   int64
			FollowerCount int64
		}{}
		err := rows.Scan(&bucket.Id, &bucket.Name, &bucket.FollowCount, &bucket.FollowerCount)
		if err != nil {
			log.Logger.Error("mysql error in writing in uResp0")
			return nil, err
		}

		uResp = append(uResp, UserFollowerItem{
			UserFollowerItem0: UserFollowerItem0{
				Id:            bucket.Id,
				Name:          bucket.Name,
				FollowCount:   bucket.FollowCount,
				FollowerCount: bucket.FollowerCount,
			},
			IsFollow: false,
		})
	}

	// 初始化
	// uResp = make(UserFollowerResp, len(uResp0))
	// for i := 0; uResp0[i].Id != 0; i++ {
	// 	uResp[i].Id = uResp0[i].Id
	// 	uResp[i].Name = uResp0[i].Name
	// 	uResp[i].FollowCount = uResp0[i].FollowCount
	// 	uResp[i].FollowerCount = uResp0[i].FollowerCount
	// 	uResp[i].IsFollow = false
	// }

	for i := 0; i < len(uResp); i++ {
		tid := u.UserId
		if err := db.MySQL.Debug().Model(&model.Follow{}).
			Where("follow_id = ? and user_id = ?", uResp[i].Id, tid).
			First(struct{}{}).Error; err != nil {
			continue
		}
		uResp[i].IsFollow = true
	}

	return uResp, nil
}
