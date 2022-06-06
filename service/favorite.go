package srv

import (
	"fmt"
	"strconv"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
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
	defer rdb.Close()
	/*
		//写入[videoID::useID]{create time}
		_, err := redis.HSet("UserLikeVideo", util.Connect(videoId, userId), time.Now().Unix()).Result()
		if err != nil {
			log.Logger.Error("set like time in redis error")
			return err
		}
	*/

	//视频点赞数计数
	err := rdb.HIncrBy("FavoriteCount", strconv.FormatInt(videoId, 10), 1).Err()
	if err != nil {
		log.Logger.Error("add like num in redis error")
		return err
	}
	//添加用户点赞的视频id
	err = rdb.ZAdd(strconv.FormatInt(userId, 10), redis.Z{Score: float64(time.Now().Unix()), Member: videoId}).Err()
	if err != nil {
		log.Logger.Error("add user favor error")
		return err
	}
	//最近点赞过用户集合
	err = rdb.SAdd("Users", strconv.FormatInt(userId, 10)).Err()
	if err != nil {
		log.Logger.Error("add user  error")
		return err
	}
	return nil
}

//取消赞
func (favorite *VideoFavorite) RemoveFavor(videoId int64, userId int64) error {
	rdb := db.Redis
	defer rdb.Close()
	/*
		err := rdb.HSet("UserLikeVideo",util.Connect(videoId,userId), "0").Err()
		if err !=nil{
			log.Logger.Error("remove like in redis error")
			return err
		}
	*/
	err := rdb.HIncrBy("FavoriteCount", strconv.FormatInt(videoId, 10), -1).Err()
	if err != nil {
		log.Logger.Error("redis error in set like num")
		return err
	}

	err = rdb.ZRem(strconv.FormatInt(userId, 10), videoId).Err()
	if err != nil {
		log.Logger.Error("redis error in list")
		return err
	}

	return nil
}

//获取点赞列表
func (favorite *VideoFavorite) FavorList(userId int64) (interface{}, error) {
	var favors []model.VideoFavorite
	result := db.MySQL.Debug().Where("user_id = ?", userId).Preload("User", "Video").Order("CreateTime desc").Find(&favors)
	fmt.Println(result)
	resp := UpdateListResp(favors)
	return resp, nil

}

func UpdateListResp(favors []model.VideoFavorite) []VideoResp {
	resp := make([]VideoResp, 0, len(favors))
	for _, favor := range favors {
		isfavor, _ := IsFavorite(favor.UserId, favor.VideoId)
		userResponse := UserResponse{
			Id:            favor.UserId,
			Name:          favor.User.Name,
			FollowCount:   favor.User.FollowCount,
			FollowerCount: favor.User.FollowerCount,
			IsFollow:      isFollowByVideoId(favor.User.Id, favor.VideoId), //未完成是否关注
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

//判断是否点赞
func IsFavorite(userId int64, videoId int64) (bool, error) {
	rdb := db.Redis
	defer rdb.Close()
	is, err := rdb.HExists("UserLikeVideo", util.Connect(videoId, userId)).Result()
	if err != nil {
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
	}
	return is, nil
}

//定时更新redis和mysql,
func RegularUpdate() {

}
func UpdateDbFavorite() error {
	//更新点赞数
	all, err := db.Redis.HGetAll("FavoriteCount").Result()
	if err != nil {
		log.Logger.Error("get all param in redis error")
		return err
	}
	for videoId, count := range all {
		if err := db.MySQL.Begin().Debug().Model(&model.Video{}).
			Where("id = ?", videoId).
			Update("favorite_count", gorm.Expr("favorite_count + ?", count)).Error; err != nil {
			db.MySQL.Begin().Rollback()
			log.Logger.Error("mysql error in updating favorite_count")
			return err
		}
	}
	//更新点赞列表
	users, err := db.Redis.SMembers("Users").Result()
	if err != nil {
		log.Logger.Error("get all param in redis error")
		return err
	}
	var favors model.VideoFavorite

	for _, userId := range users {
		videoIds, err := db.Redis.ZRange(userId, 0, -1).Result()
		if err != nil {
			log.Logger.Error("get videoId in redis error")
			return err
		}
		for _, videoId := range videoIds {
			db.MySQL.Debug().
				Model(&model.VideoFavorite{}).
				Where("video_id = ? and user_id = ?", videoId, userId).
				First(&favors)
			if favors.CreateTime == 0 {
				videoId, _ := strconv.ParseInt(videoId, 10, 64)
				userId, _ := strconv.ParseInt(userId, 10, 64)
				favors = model.VideoFavorite{
					UserId:  userId,
					VideoId: videoId,
				}
				if err := db.MySQL.Begin().Debug().Create(&favors).Error; err != nil {
					log.Logger.Error("mysql error in doing follow action")
					return err
				}
			}
		}
	}

	return nil
}
