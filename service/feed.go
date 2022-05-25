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

func (f Feed) Feed(id int64) (interface{}, error) {
	var videos []model.Video
	err := db.MySQL.Model(&model.Video{}).Order("create_time desc").Find(&videos).Error
	if err != nil {
		return nil, err
	}
	var max int64 = 0
	videoDemos := make([]VideoDemo, 0)
	for _, v := range videos {
		authorId := v.PublishId
		var u model.User
		err1 := db.MySQL.Model(&model.User{}).Where("id = ?", authorId).First(&u).Error
		if err1 != nil {
			return nil, err1
		}
		var isFollow bool
		if id == -1 {
			isFollow = false
		} else {
			var count int64
			err2 := db.MySQL.Model(&model.Follow{}).Where("user_id = ? and follow_id = ?", authorId, id).Count(&count).Error
			if err2 != nil {
				return nil, err2
			}
			if count == 0 {
				isFollow = false
			} else {
				isFollow = true
			}
		}
		userDemo := UserDemo{
			Id:            authorId,
			Name:          u.Name,
			FollowCount:   u.FollowCount,
			FollowerCount: u.FollowerCount,
			IsFollow:      isFollow,
		}
		var isFavorite bool
		if id == -1 {
			isFavorite = false
		} else {
			var count int64
			err2 := db.MySQL.Model(&model.VideoFavorite{}).Where("user_id = ? and video_id = ?", id, v.Id).Count(&count).Error
			if err2 != nil {
				return nil, err2
			}
			if count == 0 {
				isFavorite = false
			} else {
				isFavorite = true
			}
		}
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
		if max < v.CreateTime {
			max = v.CreateTime
		}
	}
	resp := FeedResp{
		NextTime:  max,
		VideoList: videoDemos,
	}
	return resp, nil
}
