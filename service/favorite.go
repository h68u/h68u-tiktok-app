package srv

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"sync"
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

//点赞操作
func (favorite *VideoFavorite) FavorAction(videoId int64, userId int64) error {
	rdb := db.Redis
	logrus.Info("videoId: ", videoId, " userId: ", userId)
	latestFlag, err := rdb.HGet(context.Background(), "FavoriteHash", util.Connect(videoId, userId)).Result()
	if err != nil {
		//log.Logger.Error("get favorite latestFlag failed", err)
		//logrus.Error("get favorite latestFlag failed", err)
		latestFlag = "0"
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
		latestFlag = "1"
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
func (v *VideoFavorite) FavorList(userId int64) (interface{}, error) {
	logrus.Info("starting favorites...")
	var m sync.Mutex
	func() {
		defer m.Unlock()
		m.Lock()
		RegularUpdate()
		}()

	logrus.Info("delete redis success")
	// 获取目标用户发布的视频
	var videos []model.VideoFavorite

	var res []VideoResp
	err := db.MySQL.Model(&model.VideoFavorite{}).Where("user_id = ?", userId).Order("create_time desc").Find(&videos).Error
	if err != nil {
		logrus.Error("mysql happen error when find video in table", err)
		return nil, err
	}
	logrus.Info("videos", videos)
	for _, video := range videos {
		var videoInTable model.Video
		//将表中的信息填到videos中，并补充其他信息
		logrus.Info("videoid: ", video.VideoId)
		err := db.MySQL.Model(&model.Video{}).Where("id = ?", video.VideoId).First(&videoInTable).Error
		if err != nil {
			logrus.Error("mysql happen error when find video in table: ", err)
			return nil, err
		}
		logrus.Info("videoInTable", videoInTable)
		
		tempvideo := VideoResp{
			Id:            videoInTable.Id,
			PlayUrl:       videoInTable.PlayUrl,
			CoverUrl:      videoInTable.CoverUrl,
			FavoriteCount: videoInTable.FavoriteCount,
			CommentCount:  videoInTable.CommentCount,
			Title:         videoInTable.Title,
		}
		tempfavorite, _ := IsFavorite(userId, video.UserId)
		tempvideo.IsFavorite = tempfavorite
		//获取作者信息
		var tempuser model.User
		err = db.MySQL.Model(&model.User{}).Where("id = ?", userId).First(&tempuser).Error
		userRes := UserResponse{
			Id:            tempuser.Id,
			Name:          tempuser.Name,
			FollowCount:   tempuser.FollowCount,
			FollowerCount: tempuser.FollowerCount,
			IsFollow:      isFollowed(userId, tempuser.Id),
		}
		tempvideo.Author = userRes
		if err != nil {
			logrus.Error("mysql happen error when query user info")
			return nil, err
		}
		res = append(res, tempvideo)
	}
	return res, nil
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
			IsFollow:      isFollowed(favor.User.Id, favor.Video.User.Id),
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

//定时更新redis和mysql,
func RegularUpdate() {
	UpdateMysql()
	DeleteRedis()
	log.Logger.Info("regular updating!")
}

func UpdateMysql() error {
	logrus.Info("Update starting1...")
	// 更新
	rdb := db.Redis
	// update mysql
	pairs, err := rdb.HGetAll(context.Background(), "FavoriteHash").Result()
	if err != nil {
		logrus.Error("get pairs failed", err)
		return err
	}
	logrus.Info("pairs", pairs)

	for pair, flag := range pairs {
		logrus.Info("Update starting3...")
		videoId, userId := util.Separate(pair)
		logrus.Info("Update starting4...")
		var favors model.VideoFavorite
		favors.UserId = userId
		favors.VideoId = videoId
		logrus.Info(userId, videoId, flag)
		if flag == "1" {
			// 更新点赞表
			// 先删除，再添加
			err = db.MySQL.Debug().Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", userId, videoId).Delete(&model.VideoFavorite{}).Error
			if err != nil {
				logrus.Error("update video favorite_count failed", err)
				return err
			}
			if err := db.MySQL.Debug().Model(&model.VideoFavorite{}).Create(&favors).Error; err != nil {
				logrus.Error("mysql error in creating video favorite")
			}
			logrus.Info("update video_favorite success")
		} else if flag == "0" {
			if err = db.MySQL.Debug().Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", userId, videoId).Delete(&model.VideoFavorite{}).Error; err != nil {
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

	}
	return nil

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
}
func DeleteRedis() error {
	//视频点赞计数可以直接删除
	err := db.Redis.Del(context.Background(), "FavoriteCount").Err()
	if err != nil {
		log.Logger.Error("delete redis error")
		return err
	}
	log.Logger.Info("delete redis count success")
	//users, err := db.Redis.SMembers(context.Background(), "Users").Result()
	//if err != nil {
	//	log.Logger.Error("get all param in redis error")
	//	return err
	//}
	//for _, userId := range users {
	//	err = db.Redis.Del(context.Background(), userId).Err()
	//	if err != nil {
	//		log.Logger.Error("delete redis error")
	//		return err
	//	}
	//}
	err = db.Redis.Del(context.Background(), "FavoriteHash").Err()
	if err != nil {
		log.Logger.Error("delete redis error")
		return err
	}
	log.Logger.Info("delete redis hash success")
	return nil
}
