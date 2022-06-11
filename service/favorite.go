package srv

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"time"
)

/*
点赞行为：
	redis:视频点赞数(hash)加1的;在用户id（key)的zset中添加点赞视频id（按照添加时间排序）;添加最近活跃过用户的id的set
取消赞行为：
redis:视频点赞数(hash)减1；在用户id（key)的zset中把点赞视频id对应的时间设置为0；添加最近活跃的用户的id的set
获取点赞列表：
	（方案1：从redis中获取点赞视频。方案2：还是先更新mysql,再删掉redis）,从mysql中获取点赞视频，按照时间排序的
更新mysql:
   更新点赞数：读取redis中视频点赞数（可以为负数），将其与mysql中的Video中的点赞数相加
   更新点赞列表：对于mysql中没有的点赞列表，根据set和zset按顺序添加。对于应该删除的点赞，根据活跃用户set和score为0的zset对数据库删除
删除redis:
	在更新mysql后全部删除，应该两者组成原子操作。
定时任务：
	以一个待定时间间隔（5分钟？）执行更新mysql和删除redis操作（应该组成原子操作）
*/
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
	/*
		//写入[videoID::useID]{create time}
		_, err := redis.HSet("UserLikeVideo", util.Connect(videoId, userId), time.Now().Unix()).Result()
		if err != nil {
			log.Logger.Error("set like time in redis error")
			return err
		}
	*/

	//视频点赞数计数
	err := rdb.HIncrBy(context.Background(),"FavoriteCount", strconv.FormatInt(videoId, 10), 1).Err()
	if err != nil {
		log.Logger.Error("add like num in redis error")
		return err
	}
	//添加用户点赞的视频id
	err = rdb.ZAdd(context.Background(),strconv.FormatInt(userId, 10), &redis.Z{Score: float64(time.Now().Unix()), Member: videoId}).Err()
	if err != nil {
		log.Logger.Error("add user favor error")
		return err
	}
	//最近活跃用户集合
	err = rdb.SAdd(context.Background(),"Users", strconv.FormatInt(userId, 10)).Err()
	if err != nil {
		log.Logger.Error("add user error")
		return err
	}
	return nil
}

//取消赞
func (favorite *VideoFavorite) RemoveFavor(videoId int64, userId int64) error {
	rdb := db.Redis
	/*
		err := rdb.HSet("UserLikeVideo",util.Connect(videoId,userId), "0").Err()
		if err !=nil{
			log.Logger.Error("remove like in redis error")
			return err
		}
	*/
	err := rdb.HIncrBy(context.Background(),"FavoriteCount", strconv.FormatInt(videoId, 10), -1).Err()
	if err != nil {
		log.Logger.Error("redis error in set like num")
		return err
	}

	err = rdb.ZAdd(context.Background(),strconv.FormatInt(userId, 10), &redis.Z{Score: float64(0), Member: videoId}).Err()
	if err != nil {
		log.Logger.Error("redis error in list")
		return err
	}
	err = rdb.SAdd(context.Background(),"Users", strconv.FormatInt(userId, 10)).Err()
	if err != nil {
		log.Logger.Error("add user  error")
		return err
	}
	return nil
}

//获取点赞列表
func (favorite *VideoFavorite) FavorList(userId int64) (interface{}, error) {
	var favors []model.VideoFavorite
	//更新数据库，删除redis
	var mu sync.Mutex
	mu.Lock()
	UpdateMysql()
	DeleteRedis()
	mu.Unlock()
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

//判断是否点赞
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

//定时更新redis和mysql,
func RegularUpdate() {
	UpdateMysql()
	DeleteRedis()
	log.Logger.Info("regular updating!")
}

func UpdateMysql() error {
	//更新点赞数
	all, err := db.Redis.HGetAll(context.Background(),"FavoriteCount").Result()
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
	users, err := db.Redis.SMembers(context.Background(), "Users").Result()
	if err != nil {
		log.Logger.Error("get all param in redis error")
		return err
	}
	var favors model.VideoFavorite

	for _, userId := range users {
		videoIds, err := db.Redis.ZRange(context.Background(),userId, 0, -1).Result()
		if err != nil {
			log.Logger.Error("get videoId in redis error")
			return err
		}
		for _, videoId := range videoIds {
			time, err := db.Redis.ZScore(context.Background(), userId, videoId).Result()
			if err != nil {
				log.Logger.Error("get time in redis error")
				return err
			}
			if time == 0 {
				videoId, _ := strconv.ParseInt(videoId, 10, 64)
				userId, _ := strconv.ParseInt(userId, 10, 64)
				favors = model.VideoFavorite{
					UserId:  userId,
					VideoId: videoId,
				}
				if err := db.MySQL.Begin().Debug().Delete(&favors).Error; err != nil {
					log.Logger.Error("mysql error in doing remove action")
					return err
				}
			}
			db.MySQL.Debug().
				Model(&model.VideoFavorite{}).
				Where("video_id = ? and user_id = ?", videoId, userId).
				First(&favors)
			if favors.CreateTime == 0 && time != 0 {
				videoId, _ := strconv.ParseInt(videoId, 10, 64)
				userId, _ := strconv.ParseInt(userId, 10, 64)
				favors = model.VideoFavorite{
					UserId:  userId,
					VideoId: videoId,
				}
				if err := db.MySQL.Begin().Debug().Create(&favors).Error; err != nil {
					log.Logger.Error("mysql error in doing favor action")
					return err
				}
			}
		}
	}
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
	err = db.Redis.Del(context.Background(),"Users").Err()
	if err != nil {
		log.Logger.Error("delete redis error")
		return err
	}
	return nil
}
