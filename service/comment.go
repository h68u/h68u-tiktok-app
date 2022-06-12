package srv

import (
	"errors"
	"gorm.io/gorm"
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
	video := model.Video{
		Id: videoId,
	}
	user, err := getUserByUserId(userId)
	if err != nil {
		return CommentResp{}, err
	}

	// 开启事务
	tx := db.MySQL.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建评论
	if err = tx.Debug().
		Model(&model.Comment{}).
		Create(&comment).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	// 评论数增加
	if err = tx.Debug().
		Model(&video).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	// 提交事务
	err = tx.Commit().Error
	return genCommentResp(comment, user), err
}

// Delete 删除评论
func (comm *Comment) Delete(userId int64, videoId int64, commentId int64) (CommentResp, error) {
	// 从数据库中找到要删除的评论
	var comment model.Comment
	video := model.Video{
		Id: videoId,
	}
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

	// 开启事务
	tx := db.MySQL.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 删除评论
	if err = tx.Debug().
		Delete(&model.Comment{}, commentId).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	// 评论数减少
	if err := tx.Debug().
		Model(&video).
		UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).
		Error; err != nil {
		tx.Rollback()
		return CommentResp{}, err
	}
	// 提交事务
	err = tx.Commit().Error
	return genCommentResp(comment, publisher), err
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

// 根据 videoId 判断是否关注了发布者
func isFollowByVideoId(userId int64, videoId int64) bool {
	var follow model.Follow
	var video model.Video
	if err := db.MySQL.Debug().
		Model(&model.Video{}).
		First(&video, videoId).
		Error; err != nil {
		return false
	}
	if err := db.MySQL.Debug().
		Model(&model.Follow{}).
		Where("follow_id = ? and user_id = ?", userId, video.PublishId).
		Find(&follow).
		Error; err != nil {
		return false
	}
	if follow.Id == 0 {
		return false
	}
	return true
}

func genCommentResp(comment model.Comment, user model.User) CommentResp {
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
		IsFollow:      isFollowByVideoId(comment.UserId, comment.VideoId),
	}
	resp.User = userResp
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
			IsFollow:      isFollowByVideoId(comment.UserId, comment.VideoId),
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
