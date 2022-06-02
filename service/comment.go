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

// Publish 发表评论
func (comm *Comment) Publish(userId int64, videoId int64, commentText string) (CommentResp, error) {
	comment := model.Comment{
		UserId:     userId,
		VideoId:    videoId,
		Content:    commentText,
		CreateTime: time.Now().Unix(),
	}
	user, err := getUserByUserId(userId)
	if err != nil {
		return CommentResp{}, err
	}
	tx := db.MySQL.Begin()
	if err = tx.Debug().
		Model(&model.Comment{}).
		Create(&comment).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	tx.Commit()
	publishId, err := getPublisherByVideoId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	return genCommentResp(comment, user, publishId), nil
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
	publisher, err := getUserByUserId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	tx := db.MySQL.Begin()
	if err = tx.Debug().
		Delete(&model.Comment{}, commentId).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	tx.Commit()
	publishId, err := getPublisherByVideoId(videoId)
	if err != nil {
		return CommentResp{}, err
	}
	return genCommentResp(comment, publisher, publishId), nil
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
	publishId, err := getPublisherByVideoId(videoId)
	if err != nil {
		return nil, err
	}
	resp := genCommentListResp(comments, publishId)
	return resp, nil
}

// 根据 userId 获得 user
func getUserByUserId(userId int64) (model.User, error) {
	var user model.User
	if err := db.MySQL.Debug().
		Model(&model.User{}).
		First(&user, userId).
		Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

// 根据 videoId 获得 publishId
func getPublisherByVideoId(videoId int64) (int64, error) {
	var video model.Video
	if err := db.MySQL.Debug().
		Model(&model.Video{}).
		First(&video, videoId).
		Error; err != nil {
		return 0, err
	}
	return video.PublishId, nil
}

// 判断 userId 是否关注了 publishId
func isFollow(userId int64, publishId int64) bool {
	var follow model.Follow
	if err := db.MySQL.Debug().
		Model(&model.Follow{}).
		Where("follow_id = ? and user_id = ?", publishId, userId).
		Find(&follow).
		Error; err != nil {
		return false
	}
	return true
}

func genCommentResp(comment model.Comment, user model.User, publishId int64) CommentResp {
	resp := CommentResp{
		Id:         comment.Id,
		Content:    comment.Content,
		CreateDate: time.Unix(comment.CreateTime, 0).Format("2006-01-02 03:04:05 PM"),
		User:       UserResp{},
	}
	userResp := UserResp{
		Id:            user.Id,
		Name:          user.Name,
		FollowCount:   user.FollowerCount,
		FollowerCount: user.FollowCount,
		IsFollow:      isFollow(comment.UserId, publishId),
	}
	resp.User = userResp
	return resp
}

func genCommentListResp(comments []model.Comment, publishId int64) []CommentResp {
	resp := make([]CommentResp, 0, len(comments))
	for _, comment := range comments {
		userResp := UserResp{
			Id:            comment.User.Id,
			Name:          comment.User.Name,
			FollowCount:   comment.User.FollowCount,
			FollowerCount: comment.User.FollowerCount,
			IsFollow:      isFollow(comment.UserId, publishId),
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
