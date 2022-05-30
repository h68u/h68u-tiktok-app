package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"
	"fmt"

)

type VideoFavorite struct{}



type VideoDemo struct {
	Id            int64    `json:"id"`
	Author        UserDemo `json:"author"`
	PlayUrl       string   `json:"play_url"`
	CoverUrl      string   `json:"cover_url"`
	FavoriteCount int64    `json:"favorite_count"`
	CommentCount  int64    `json:"comment_count"`
	IsFavorite    bool     `json:"is_favorite"`
	Title         string   `json:"title"`
}

type UserDemo struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	IsFollow      bool   `json:"is_follow"`
}
redis := db.Redis

//后续设置context？需要加并行?
//点赞操作
func (favorite *VideoFavorite) SetFavor(videoId int64,userId int64) (error){
    redis := db.Redis
	defer redis.Close()
    //写入[videoID::useID]{create time}
	res,err:=redis.HSet("UserLikeVideo",util.Connect(videoId,userId),time.Now().Unix())
	if err != nil{
		log.Logger.Error("set like time in redis error")
		return err
	}
	//视频点赞数计数
	_ , err := redis.hincrby("FavoriteCount",videoId,1)
	if err != nil{
		log.Logger.Error("add like num in redis error")
		return err
	}
	return 
}

//取消赞
func (favorite *VideoFavorite)RemoveFavor(videoId int64,userId int64)(error){
	
	defer redis.Close()
	_, err := redis.hdel("UserLikeVideo",util.Connect(videoId,userId))
	if err !=nil{
		log.Logger.Error("remove like in redis error")
		return err
	}
	_, err1:= redis.hget("FavoriteCount",videoId) 
	if err1 != nil{
		log.Logger.Error("get num in redis error")
		return err1

	}
	if count >0{
		_ . err2 :=redis.hincrby("FavoriteCount",videoId,-1)
		if err2 !=nil {
			log,Logger.Error("redis error in set like num")
			return err2
		}
	}
	return 	
}

//获取点赞列表
func (favorite *Favorite)FavorList(userId int64)([]FavorListResp,error) {
	var favors []model.VideoFavorite
	result := db.MySQL.Debug().Where("user_id = ?", userId).Preload("User";"Video").Order("CreateTime desc").Find(&favors)
	fmt.Println(result)
	resp := UpdateListResp(favors)
	return resp,nil

}

func UpdateListResp(favors []model.VideoFavorite) ([]FavorListResp){
	resp :=make([]FavorListResp,0,len(favors))
	for _, favor := range favors {
		UserDemo := UserResp{
			Id:			   favor.UserId,
			Name:		   favor.User.Name,
			FollowCount:   favor.User.FollowCount,
			FollowerCount: favor.User.FollowerCount,
			IsFollow:      isFollow(favor.User.Id),    //未完成是否关注
		}
		VideoDemo := VideoDemo{
		Id            	favor.VideoId,
		Author       	favor.UserDemo,
		PlayUrl       	favor.Video.PlayUrl,
		CoverUrl      	favor.Video.CoverUrl,
		FavoriteCount 	favor.Video.FavoriteCount,
		CommentCount    favor.Video.CommentCount,
		IsFavorite      IsFavorite(favor.VideoId, favor.UserId)
		Title           favor.Video.Title,
		}
		resp = append(resp,VideoDemo)		
	}
	return resp
}

//判断是否点赞
func IsFavorite(videoId int64,userId int64)(bool,error)  {
	defer redis.Close()
	is ,err := db.redis.hexists("UserLikeVideo",util.Connect(videoId,userId))
	if err != nil{
		log.Logger.Error("isfavorite can not be known ")
		return nil , err
	}
	return is,nil
}

//定时更新redis和mysql,
func RegularUpdate(){

}