package srv

import (
	"bytes"
	"errors"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"tikapp/common/db"
	"tikapp/common/log"
	"tikapp/common/model"
	"tikapp/common/oss"
	"time"
)

type Video struct{}

const BucketName = "tiktok-video11"

func (v Video) PublishAction(data *multipart.FileHeader, title string, publishId int64) error {
	//oss.CreateBucket(BucketName)
	// 获取文件
	file, err := data.Open()
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)
	// 判断是否为视频
	checkFile, err := data.Open()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, checkFile); err != nil {
		logrus.Error("copy file error", err)
		return err
	}
	if filetype.IsVideo(buf.Bytes()) == false {
		logrus.Error("file is not video")
		return errors.New("not a video")
	}
	checkFile.Close()
	// 存储到oss
	log.Logger.Info("start to upload video to oss, file type: ")
	ok, err := oss.UploadVideoToOss(BucketName, data.Filename, file)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("upload video error")
	}

	// 获取url 存储到数据库
	videoUrl, imgUrl, err := oss.GetOssVideoUrlAndImgUrl(BucketName, data.Filename)
	if err != nil {
		return err
	}
	video := model.Video{
		PublishId:     publishId,
		PlayUrl:       videoUrl,
		CoverUrl:      imgUrl,
		FavoriteCount: 0,
		CommentCount:  0,
		Title:         title,
		CreateTime:    time.Now().Unix(),
	}
	logrus.Info("play_url: ", videoUrl)
	err = db.MySQL.Model(&model.Video{}).Create(&video).Error
	if err != nil {
		return err
	}
	logrus.Info("inset video to mysql success")
	return nil
}

func (v Video) PublishList(myUserID, targetUserID int64) ([]VideoDemo, error) {
	// 获取目标用户发布的视频
	var videos []VideoDemo
	var videoInTable []model.Video
	err := db.MySQL.Model(&model.Video{}).Where("publish_id = ?", targetUserID).Find(&videoInTable).Error
	if err != nil {
		logrus.Error("mysql happen error when find video in table", err)
		return nil, err
	}

	for _, v := range videoInTable { //将表中的信息填到videos中，并补充其他信息
		var video VideoDemo
		video.Id = v.Id
		video.Title = v.Title
		video.PlayUrl = v.PlayUrl
		video.CoverUrl = v.CoverUrl
		video.FavoriteCount = v.FavoriteCount
		video.CommentCount = v.CommentCount
		//video.CreateTime = v.CreateTime 返回体中没有create_time字段
		//video.PublishId = v.PublishId 返回体中没有publish_id字段
		video.IsFavorite, err = IsFavorite(myUserID, v.Id)
		if err != nil {
			logrus.Error("mysql happen error when query favorite")
			return nil, err
		}
		u := User{}
		video.Author, err = u.Info(myUserID, targetUserID)
		if err != nil {
			logrus.Error("mysql happen error when query user info")
			return nil, err
		}
		videos = append(videos, video)
	}
	return videos, nil
}
