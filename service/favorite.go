package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/util/stringconnect"
	"time"

)

type Favorite struct{}

//需要加并行

/* type FavoriteReq struct {
	UserId      int64    `json:"user_id"`
	Token       string   `json:"token` 
	VideoId     int64    `json:"video_id`
	Actiontype  int64    `json:"action_type"`
} */



//后续设置context？
//点赞操作
func (favor *Favorite) SetFavor(videoId int64,userId int64) (error){
    redis := db.Redis
	defer redis.Close()
    //写入[videoID::useID]{create time}
	res,err:=redis.HSet("UserLikeVideo",stringconnect(videoId,userId),time.Now().Unix())
	if err != nil{

	}
	//视频点赞数计数
	_ , err := redis.hincrby("FavoriteCount",videoId,1)
	if err != nil{

	}
}

//取消赞
func (favor *Favorite)RemoveFavor(videoId int64,userId int64)(error){
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
func (favor *Favorite)FavorList(userId int64){
	


}

//定时更新redis和mysql
func RegularUpdate(){

}