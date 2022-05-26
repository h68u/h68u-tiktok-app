package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/model"
	"time"
)

type Comment struct{}

type CommentResp struct {
	Id         int64    `json:"id"`
	Content    string   `json:"content"`
	CreateDate string   `json:"create_date"`
	User       UserResp `json:"user"`
}

type UserResp struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
}

func (comm *Comment) Publish(userId int64, videoId int64, commentText string) (CommentResp, error) {
	comment := model.Comment{
		UserId:     userId,
		VideoId:    videoId,
		Content:    commentText,
		CreateTime: time.Now().Unix(),
	}
	var user model.User
	err := db.MySQL.Debug().Model(&model.User{}).Find(&user, userId).Error
	if err != nil {
		return CommentResp{}, err
	}
	err = db.MySQL.Debug().Model(&model.Comment{}).Create(&comment).Error
	if err != nil {
		return CommentResp{}, err
	}

	resp := CommentResp{
		Id:         comment.Id,
		Content:    comment.Content,
		CreateDate: time.Unix(comment.CreateTime, 0).Format("2006-01-02 03:04:05 PM"),
		User:       UserResp{},
	}
	resp.User = UserResp(user)
	return resp, nil
}

var ErrPermit = errors.New("permission is not allowed")

func (comm *Comment) Delete(userId int64, videoId int64, commentId int64) (CommentResp, error) {
	var comment model.Comment
	err := db.MySQL.Debug().Model(&model.Comment{}).Find(&comment, commentId).Error
	if err != nil {
		return CommentResp{}, err
	}
	if comment.UserId != userId {
		return CommentResp{}, ErrPermit
	}
	if comment.VideoId != videoId {

	}
	db.MySQL.Debug().Delete(&model.Comment{}, commentId)
	return CommentResp{}, nil
}
