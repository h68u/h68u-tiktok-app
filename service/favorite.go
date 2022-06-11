package srv

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"
)

type VideoFavorite struct{}

type VideoResp struct {
	Id            int64        `json:"id"`
	Author        UserResponse `json:"author"`
	PlayUrl       string       `json:"play_url"`
	CoverUrl      string       `json:"cover_url"`
	FavoriteCount int64        `json:"favorite_count"`
	CommentCount  int64        `json:"comment_count"`
	IsFavorite    bool         `json:"is_favorite"`
	Title         string       `json:"title"`
}

type UserResponse struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	IsFollow      bool   `json:"is_follow"`
}

//后续设置context？需要加并行?
//点赞操作
func (favorite *VideoFavorite) FavorAction(videoId int64, userId int64) error {
	rdb := db.Redis
	logrus.Info("videoId: ", videoId, " userId: ", userId)
	/*
		//写入[videoID::useID]{create time}
		_, err := redis.HSet("UserLikeVideo", util.Connect(videoId, userId), time.Now().Unix()).Result()
		if err != nil {
			log.Logger.Error("set like time in redis error")
			return err
		}
	*/
	latestFlag, err := rdb.HGet(context.Background(), "FavoriteHash", util.Connect(videoId, userId)).Result()
	if err != nil {
		log.Logger.Error("get favorite latestFlag failed")
		return err
	}
	logrus.Info("latestFlag: ", latestFlag)
	if latestFlag != "1" {
		_, err := rdb.HSet(context.Background(), "FavoriteHash", util.Connect(videoId, userId), 1).Result()
		if err != nil {
			logrus.Error("set hash error: ", err)
			return err
		}
		logrus.Info("videoid: ", videoId, " userId: ", userId, " liked success", " latestFlag: ", latestFlag)
		//视频点赞数计数
		err = rdb.HIncrBy(context.Background(), "FavoriteCount", strconv.FormatInt(videoId, 10), 1).Err()
		if err != nil {
			log.Logger.Error("add like num in redis error")
			return err
		}
		logrus.Info("increase like num success")
	}
	return nil
}

//取消赞
func (favorite *VideoFavorite) RemoveFavor(videoId int64, userId int64) error {
	rdb := db.Redis
	latestFlag, err := rdb.HGet(context.Background(), "FavoriteHash", util.Connect(videoId, userId)).Result()
	if err != nil {
		log.Logger.Error("get favorite latestFlag failed")
		return err
	}
	logrus.Info("latestFlag: ", latestFlag)
	if latestFlag != "0" {
		_, err := rdb.HSet(context.Background(), "FavoriteHash", util.Connect(videoId, userId), 0).Result()
		if err != nil {
			logrus.Error("set hash error: ", err)
			return err
		}
		logrus.Info("videoid: ", videoId, " userId: ", userId, " unliked success")
		//视频点赞数计数
		err = rdb.HIncrBy(context.Background(), "FavoriteCount", strconv.FormatInt(videoId, 10), -1).Err()
		if err != nil {
			log.Logger.Error("decrease like num in redis error")
			return err
		}
		logrus.Info("decrease like num success")
	}
	return nil
}

//获取点赞列表
func (favorite *VideoFavorite) FavorList(userId int64) (interface{}, error) {
	logrus.Info("starting favorites...")
	var favors []model.VideoFavorite
	//更新数据库，删除redis
	//var mu sync.Mutex
	//mu.Lock()
	err := UpdateMysql()
	if err != nil {
		logrus.Error("update favorite failed", err)
		return nil, err
	}
	DeleteRedis()
	//mu.Unlock()
	logrus.Info("favors:", favors)
	resp := UpdateListResp(favors)
	return resp, nil
}
func UpdateListResp(favors []model.VideoFavorite) []VideoResp {
	resp := make([]VideoResp, 0, len(favors))
	for _, favor := range favors {
		logrus.Info("favor: ", favor)
		isfavor, _ := IsFavorite(favor.UserId, favor.VideoId)
		userResponse := UserResponse{
			Id:            favor.UserId,
			Name:          favor.User.Name,
			FollowCount:   favor.User.FollowCount,
			FollowerCount: favor.User.FollowerCount,
			IsFollow:      isFollowed(favor.User.Id, favor.Video.User.Id), //未完成是否关注
		}
		videoResp := VideoResp{
			Id:            favor.VideoId,
			Author:        userResponse,
			PlayUrl:       favor.Video.PlayUrl,
			CoverUrl:      favor.Video.CoverUrl,
			FavoriteCount: favor.Video.FavoriteCount,
			CommentCount:  favor.Video.CommentCount,
			IsFavorite:    isfavor,
			Title:         favor.Video.Title,
		}
		resp = append(resp, videoResp)
	}
	return resp
}

// IsFavorite 判断是否点赞
func IsFavorite(userId int64, videoId int64) (bool, error) {
	// rdb := db.Redis
	log.Logger.Error("isfavorite can not be known ")
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

// RegularUpdate 定时更新redis和mysql,
func RegularUpdate() error {
	var mu sync.Mutex
	go func() {
		time.Sleep(time.Minute * 10)
		mu.Lock()
		defer mu.Unlock()
		logrus.Info("update mysql and redis")
		UpdateMysql()
		DeleteRedis()
	}()
	return nil
}
func UpdateMysql() error {
	// 更新
	rdb := db.Redis
	// update mysql
	pairs, err := rdb.HGetAll(context.Background(), "VideoHash").Result()
	if err != nil {
		logrus.Error("get pairs failed", err)
		return err
	}

	for pair, flag := range pairs {
		videoId, userId := util.Separate(pair)
		var favors model.VideoFavorite
		favors.UserId = userId
		favors.VideoId = videoId
		if flag == "1" {
			// 更新点赞表
			// 先删除，再添加
			err = db.MySQL.Debug().Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", videoId, userId).Delete(&VideoFavorite{}).Error
			if err != nil {
				logrus.Error("update video favorite_count failed", err)
				return err
			}
			if err := db.MySQL.Debug().Model(&model.VideoFavorite{}).Create(&favors).Error; err != nil {
				logrus.Error("mysql error in creating video favorite")
			}
		} else if flag == "0" {
			if err = db.MySQL.Debug().Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", videoId, userId).Delete(&VideoFavorite{}).Error; err != nil {
				logrus.Error("mysql error in deleting video favorite")
			}
		}
		// 更新视频点赞数
		delta, err := rdb.HGet(context.Background(), "FavoriteCount", strconv.FormatInt(videoId, 10)).Result()
		if err != nil {
			logrus.Error("get delta failed", err)
		}
		logrus.Info("delta: ", delta)
		if err := db.MySQL.Debug().Model(&model.Video{}).
			Where("id = ?", videoId).
			Update("favorite_count", gorm.Expr("favorite_count + ?", delta)).Error; err != nil {
			logrus.Error("mysql error in updating favorite_count")
			return err
		}
		return nil
	}

	//count := rdb.BitCount{Start: 0, End: -1}
	//hashMap, err := rdb.HGetAll(context.Background(), "FavoriteCount").Result()
	//logrus.Info("hashMap:", hashMap)
	//if err != nil {
	//	log.Logger.Error("get all param in redis error")
	//	return err
	//}
	//for videoId, count := range hashMap {
	//	logrus.Info("videoId:", videoId, " count:", count)
	//	if err := db.MySQL.Debug().Model(&model.Video{}).
	//		Where("id = ?", videoId).
	//		Update("favorite_count", gorm.Expr("favorite_count + ?", count)).Error; err != nil {
	//		//db.MySQL.Begin().Rollback()
	//		logrus.Error("mysql error in updating favorite_count")
	//		return err
	//	}
	//}
	////更新点赞列表
	//users, err := db.Redis.SMembers(context.Background(), "Users").Result()
	//logrus.Info("users: ", users)
	//if err != nil {
	//	log.Logger.Error("get all param in redis error")
	//	return err
	//}
	//var favors model.VideoFavorite
	//
	//for _, userId := range users {
	//	videoIds, err := db.Redis.ZRange(context.Background(), userId, 0, -1).Result()
	//	if err != nil {
	//		logrus.Error("get videoId in redis error")
	//		return err
	//	}
	//	for _, videoId := range videoIds {
	//		videoId, _ := strconv.ParseInt(videoId, 10, 64)
	//		userId, _ := strconv.ParseInt(userId, 10, 64)
	//		favors.UserId = userId
	//		favors.VideoId = videoId
	//		if err := db.MySQL.Debug().Model(&model.VideoFavorite{}).Create(&favors).Error; err != nil {
	//			logrus.Error("mysql error in creating video favorite")
	//		}
	//
	//		//db.MySQL.Debug().
	//		//	Model(&model.VideoFavorite{}).
	//		//	Where("video_id = ? and user_id = ?", videoId, userId).
	//		//	First(&favors)
	//		//if favors.CreateTime == 0 {
	//		//	videoId, _ := strconv.ParseInt(videoId, 10, 64)
	//		//	userId, _ := strconv.ParseInt(userId, 10, 64)
	//		//	favors = model.VideoFavorite{
	//		//		UserId:  userId,
	//		//		VideoId: videoId,
	//		//	}
	//		//	logrus.Info("favors: ", favors)
	//		//	if err := db.MySQL.Debug().Create(&favors).Error; err != nil {
	//		//		log.Logger.Error("mysql error in doing follow action")
	//		//		return err
	//		//	}
	//		//}
	//	}
	//}
	return nil
}
func DeleteRedis() error {
	//视频点赞计数可以直接删除
	err := db.Redis.Del(context.Background(), "FavoriteCount").Err()
	if err != nil {
		log.Logger.Error("delete redis error")
		return err
	}
	users, err := db.Redis.SMembers(context.Background(), "Users").Result()
	if err != nil {
		log.Logger.Error("get all param in redis error")
		return err
	}
	for _, userId := range users {
		err = db.Redis.Del(context.Background(), userId).Err()
		if err != nil {
			log.Logger.Error("delete redis error")
			return err
		}
	}
	err = db.Redis.Del(context.Background(), "Users").Err()
	if err != nil {
		log.Logger.Error("delete redis error")
		return err
	}
	return nil
}
