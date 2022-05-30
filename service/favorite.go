package srv

import (
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"
	"fmt"
	"strconv"

)

type VideoFavorite struct{}



type VideoResp struct {
	Id            int64    `json:"id"`
	Author  UserResponse `json:"author"`
	PlayUrl       string   `json:"play_url"`
	CoverUrl      string   `json:"cover_url"`
	FavoriteCount int64    `json:"favorite_count"`
	CommentCount  int64    `json:"comment_count"`
	IsFavorite    bool     `json:"is_favorite"`
	Title         string   `json:"title"`
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
func (favorite *VideoFavorite) SetFavor(videoId int64,userId int64) (error){
	redis := db.Redis
	defer redis.Close()
    //写入[videoID::useID]{create time}
	_, err := redis.HSet("UserLikeVideo",util.Connect(videoId,userId),time.Now().Unix()).Result()
	if err != nil{
		log.Logger.Error("set like time in redis error")
		return err
	}
	//视频点赞数计数
	_, err = redis.HIncrBy("FavoriteCount",string(videoId),1).Result()
	if err != nil{
		log.Logger.Error("add like num in redis error")
		return err
	}
	return nil
}

//取消赞
func (favorite *VideoFavorite)RemoveFavor(videoId int64,userId int64)(error){
	redis := db.Redis
	defer redis.Close()
	_, err := redis.HDel("UserLikeVideo",util.Connect(videoId,userId)).Result()
	if err !=nil{
		log.Logger.Error("remove like in redis error")
		return err
	}
	count, err:= redis.HGet("FavoriteCount",string(videoId)).Result()
	if err != nil{
		log.Logger.Error("get num in redis error")
		return err
	}
	coun, err := strconv.Atoi(count)
	if err != nil{
		log.Logger.Error("convert int error")
		return err
	}
	if coun > 0{
		_, err :=redis.HIncrBy("FavoriteCount",string(videoId),-1).Result()
		if err !=nil {
			log.Logger.Error("redis error in set like num")
			return err
		}
	}
	return 	nil
}

//获取点赞列表
func (favorite *VideoFavorite)FavorList(userId int64)(interface{},error) {
	var favors []model.VideoFavorite
	result := db.MySQL.Debug().Where("user_id = ?", userId).Preload("User" ,"Video").Order("CreateTime desc").Find(&favors)
	fmt.Println(result)
	resp := UpdateListResp(favors)
	return resp,nil

}

func UpdateListResp(favors []model.VideoFavorite) ([]VideoResp){
	resp := make([]VideoResp, 0, len(favors))
	for _, favor := range favors {
		isfavor, _ := IsFavorite(favor.UserId, favor.VideoId)
		userResponse := UserResponse{
			Id:			   favor.UserId,
			Name:		   favor.User.Name,
			FollowCount:   favor.User.FollowCount,
			FollowerCount: favor.User.FollowerCount,
			IsFollow:      isFollow(favor.User.Id),    //未完成是否关注
		}
		videoResp := VideoResp{
		Id:            	 favor.VideoId,
		Author:       	 userResponse,
		PlayUrl:       	 favor.Video.PlayUrl,
		CoverUrl:      	 favor.Video.CoverUrl,
		FavoriteCount: 	 favor.Video.FavoriteCount,
		CommentCount:    favor.Video.CommentCount,
		IsFavorite:      isfavor,
		Title:           favor.Video.Title,
		}
		resp = append(resp,videoResp)		
	}
	return resp
}

//判断是否点赞
func IsFavorite(userId int64, videoId int64)(bool,error)  {
	redis := db.Redis
	defer redis.Close()
	is ,err := redis.HExists("UserLikeVideo",util.Connect(videoId,userId)).Result()
	if err != nil{
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
	return is,nil
}

//定时更新redis和mysql,
func RegularUpdate(){

}
