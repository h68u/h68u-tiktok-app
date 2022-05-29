package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util"
	"time"

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


//后续设置context？需要加并行?
//点赞操作
func (favorite *VideoFavorite) SetFavor(videoId int64,userId int64) (error){
    redis := db.Redis
	defer redis.Close()
    //写入[videoID::useID]{create time}
	res,err:=redis.HSet("UserLikeVideo",util.connect(videoId,userId),time.Now().Unix())
	if err != nil{

	}
	//视频点赞数计数
	_ , err := redis.hincrby("FavoriteCount",videoId,1)
	if err != nil{

	}
}

//取消赞
func (favorite *VideoFavorite)RemoveFavor(videoId int64,userId int64)(error){
	redis := db.Redis
	defer redis.Close()
	_,err := redis.hdel("UserLikeVideo",stringconnect(videoId,userId))
	if err !=nil{

	}
	count , err1:= redis.hget("FavoriteCount",videoId) 
	if err1 != nil{

	}
	if count >0{
		_ . err2 :=redis.hincrby("FavoriteCount",videoId,-1)
		if err2 !=nil {

		}
	}	
}

//获取点赞列表
func (favorite *Favorite)FavorList(userId int64)([]FavorListResp,error) {
	var favors []model.VideoFavorite
	result := db.MySQL.Debug().Where("user_id = ?", userId).Preload("User";"Video").Order("CreateTime desc").Find(&favors)
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
		IsFavorite      favor.Video.CreateTime,
		Title           favor.Video.Title,
		}
		resp = append(resp,VideoDemo)		
	}
	return resp
}

//定时更新redis和mysql
func RegularUpdate(){

}