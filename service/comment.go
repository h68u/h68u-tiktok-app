package srv

import (
	"errors"
	"tikapp/common/db"
	"tikapp/common/model"
	"time"
)

var ErrPermit = errors.New("permission is not allowed")

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
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	IsFollow      bool   `json:"is_follow"`
}

// 根据 videoId 获得视频发布者 user
func getPublisherByVideoId(videoId int64) (model.User, error) {
	var video model.Video
	err := db.MySQL.Debug().Model(&model.Video{}).First(&video, videoId).Error
	if err != nil {
		return model.User{}, err
	}
	var user model.User
	err = db.MySQL.Debug().Model(&model.User{}).First(&user, video.PublishId).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// todo 根据 userId 判断是否关注
func isFollow(userId int64) bool {
	return false
}

func generateResp(comment model.Comment, user model.User) CommentResp {
	resp := CommentResp{
		Id:         comment.Id,
		Content:    comment.Content,
		CreateDate: time.Unix(comment.CreateTime, 0).Format("2006-01-02 03:04:05 PM"),
		User:       UserResp{},
	}
	publisher := UserResp{
		Id:            user.Id,
		Name:          user.Name,
		FollowCount:   user.FollowerCount,
		FollowerCount: user.FollowCount,
		IsFollow:      isFollow(user.Id),
	}
	resp.User = publisher
	return resp
}

// Publish 发表评论
func (comm *Comment) Publish(userId int64, videoId int64, commentText string) (CommentResp, error) {
	comment := model.Comment{
		UserId:     userId,
		VideoId:    videoId,
		Content:    commentText,
		CreateTime: time.Now().Unix(),
	}
	publisher, err := getPublisherByVideoId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	err = db.MySQL.Debug().Model(&model.Comment{}).Create(&comment).Error
	if err != nil {
		return CommentResp{}, err
	}
	return generateResp(comment, publisher), nil
}

// Delete 删除评论
func (comm *Comment) Delete(userId int64, videoId int64, commentId int64) (CommentResp, error) {
	var comment model.Comment
	err := db.MySQL.Debug().Model(&model.Comment{}).First(&comment, commentId).Error
	if err != nil {
		return CommentResp{}, err
	}
	if comment.UserId != userId {
		return CommentResp{}, ErrPermit
	}
	publisher, err := getPublisherByVideoId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	db.MySQL.Debug().Delete(&model.Comment{}, commentId)
	return generateResp(comment, publisher), nil
}
