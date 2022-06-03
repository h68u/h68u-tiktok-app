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
	UserId     int64
	Token      string
	ToUserId   int64
	ActionType int32
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
func (r Relation) FollowList(u *UserFollowerReq, visitorId int64) (UserFollowerResp, error) {
	var uResp UserFollowerResp

	
	// visiterId 指代的是 token 的主人的 id
	// 这里不排除 用户A 去看 用户B 的关注列表，所以需要在 用户B 所关注的人找出 用户A 也关注的对象？
	// 在这里假定 用户A 为 token 的主人，而 用户B 为请求中的 user_id 的主人


	// 获取 用户B 所关注的人的信息
	rows, err := db.MySQL.Debug().Raw(sqlUserFollower, u.UserId).Rows()
	if err != nil {
		log.Logger.Error("mysql error in get follower list")
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

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

	// 将 用户A/B 共同关注的人找出
	for i := range uResp {
		if isFollowed(visitorId, uResp[i].Id) {
			uResp[i].IsFollow = true
		}
	}

	return uResp, nil
}

// isFollowed 检查给定 h: host 是否关注了 u: user (给定的user)
func isFollowed(h, u int64) bool {
	if err := db.MySQL.Debug().
		Where("follow_id = ? and user_id = ?", h, u).
		First(&model.Follow{}).Error; err != nil {
		return false
	}

	return true
}

// FollowerList 获取给定用户的粉丝列表
func FollowerList(userId int64) ([]UserDemo, error) {
	var ans []UserDemo
	//var followInTable model.Follow

	//获取所有粉丝
	rows, err := db.MySQL.Debug().Table("follow").
		Select("follow_id").
		Where("user_id = ?", userId).
		Rows()
	if err != nil {
		log.Logger.Error("mysql error in get follower list")
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	//逐个处理每个粉丝信息
	for rows.Next() {
		var u User            // 因为Info方法是User结构体的方法
		var fansId int64      //要从粉丝表中获取的粉丝id
		var fansInfo UserDemo //粉丝信息

		err := rows.Scan(&fansId)
		if err != nil {
			log.Logger.Error("mysql error in writing in fansId")
			return nil, err
		}

		fansInfo, err = u.Info(userId, fansId)
		if err != nil {
			log.Logger.Error("get userInfo err")
			return nil, err
		}

		ans = append(ans, fansInfo)

	}

	return ans, nil
}
