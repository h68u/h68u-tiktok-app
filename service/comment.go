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

var tx = db.MySQL.Begin()

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
	if err = tx.Debug().
		Model(&model.Comment{}).
		Create(&comment).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	return genCommentResp(comment, publisher), nil
}

// Delete 删除评论
func (comm *Comment) Delete(userId int64, videoId int64, commentId int64) (CommentResp, error) {
	// 从数据库中找到要删除的评论
	var comment model.Comment
	if err := db.MySQL.Debug().
		Model(&model.Comment{}).
		First(&comment, commentId).
		Error; err != nil {
		return CommentResp{}, err
	}
	// 验证执行删除操作的当前用户是否是该评论的发布者
	if comment.UserId != userId {
		return CommentResp{}, ErrPermit
	}
	// 获取这条评论所属视频id,用于之后查询视频信息
	publisher, err := getPublisherByVideoId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	err = tx.Debug().Delete(&model.Comment{}, commentId).Error
	if err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	return genCommentResp(comment, publisher), nil
}

func (comm *Comment) List(videoId int64) ([]CommentResp, error) {
	var comments []model.Comment
	if err := db.MySQL.Debug().
		Where("video_id = ?", videoId).
		Preload("User").
		Order("create_time desc").
		Find(&comments).Error; err != nil {
		return nil, err
	}
	resp := genCommentListResp(comments)
	return resp, nil
}

// 根据 videoId 获得视频发布者 user
func getPublisherByVideoId(videoId int64) (model.User, error) {
	var video model.Video
	if err := db.MySQL.Debug().
		Model(&model.Video{}).
		First(&video, videoId).
		Error; err != nil {
		return model.User{}, err
	}
	var user model.User
	if err := db.MySQL.Debug().
		Model(&model.User{}).
		First(&user, video.PublishId).
		Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

// todo 根据 userId 判断是否关注
func isFollow(userId int64) bool {
	return false
}

func genCommentResp(comment model.Comment, user model.User) CommentResp {
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

func genCommentListResp(comments []model.Comment) []CommentResp {
	resp := make([]CommentResp, 0, len(comments))
	for _, comment := range comments {
		userResp := UserResp{
			Id:            comment.User.Id,
			Name:          comment.User.Name,
			FollowCount:   comment.User.FollowCount,
			FollowerCount: comment.User.FollowerCount,
			IsFollow:      isFollow(comment.User.Id),
		}
		commentResp := CommentResp{
			Id:         comment.Id,
			Content:    comment.Content,
			CreateDate: time.Unix(comment.CreateTime, 0).Format("2006-01-02 03:04:05 PM"),
			User:       userResp,
		}
		resp = append(resp, commentResp)
	}

	return resp
}
