package srv

import (
	"tikapp/common/db"
	"tikapp/common/model"
)

type Feed struct{}

type FeedResp struct {
	NextTime  int64       `json:"next_time"`
	VideoList []VideoDemo `json:"video_list"`
}

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

// Feed 获取视频列表
// id 若为-1，表示没有获取到用户id
// lastTime 值0时不限制；限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
// nextTime 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
func (f Feed) Feed(id int64, lastTime int64) (interface{}, error) {
	// 目前：取出最新视频
	// TODO：控制视频数量，应该不是简单limit，需要保证视频一直可以刷下去，可能涉及并发

	var videos []model.Video
	if lastTime != 0 {
		// TODO：时间单位确定
		err := db.MySQL.Model(&model.Video{}).
			Where("created_at < ?", lastTime).
			Order("created_at desc").
			Limit(30).Find(&videos).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := db.MySQL.Model(&model.Video{}).
			Order("create_time desc").
			Limit(30).Find(&videos).Error
		if err != nil {
			return nil, err
		}
	}

	var nextTime int64 = 0
	videoDemos := make([]VideoDemo, 0)
	for _, v := range videos {
		// 获取作者信息
		authorId := v.PublishId
		var u model.User
		err := db.MySQL.Model(&model.User{}).Where("id = ?", authorId).First(&u).Error
		if err != nil {
			return nil, err
		}

		// 用户是否关注了
		var isFollow bool
		if id == -1 {
			isFollow = false
		} else {
			var count int64
			err := db.MySQL.Model(&model.Follow{}).
				Where("user_id = ? and follow_id = ?", authorId, id).
				Count(&count).Error
			if err != nil {
				return nil, err
			}
			if count == 0 {
				isFollow = false
			} else {
				isFollow = true
			}
		}

		// 视频作者基本信息
		userDemo := UserDemo{
			Id:            authorId,
			Name:          u.Name,
			FollowCount:   u.FollowCount,
			FollowerCount: u.FollowerCount,
			IsFollow:      isFollow, // 用户是否关注了作者
		}
		// 用户是否点赞了
		var isFavorite bool
		if id == -1 {
			isFavorite = false
		} else {
			var count int64
			err := db.MySQL.Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", id, v.Id).Count(&count).Error
			if err != nil {
				return nil, err
			}
			if count == 0 {
				isFavorite = false
			} else {
				isFavorite = true
			}
		}

		// 视频信息
		videoDemos = append(videoDemos, VideoDemo{
			Id:            v.Id,
			Author:        userDemo,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			IsFavorite:    isFavorite,
			Title:         v.Title,
		})
		if nextTime < v.CreateTime {
			nextTime = v.CreateTime
		}
	}

	resp := FeedResp{
		NextTime:  nextTime,
		VideoList: videoDemos,
	}
	return resp, nil
}
