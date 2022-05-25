package srv

import (
	"errors"
	"mime/multipart"
	"tikapp/common/db"
	"tikapp/common/model"
	"tikapp/common/oss"
	"time"
)

type VideoPublishReq struct {
	Token string               `json:"token"`
	Title string               `json:"title"`
	Data  multipart.FileHeader `json:"data"`
}

type Video struct{}

const BucketName = "tiktok-video11"

func (v Video) PublishAction(data *multipart.FileHeader, title string, publishId int64) error {
	//oss.CreateBucket(BucketName)
	file, err := data.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	ok, err := oss.UploadVideoToOss(BucketName, data.Filename, file)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("upload video error")
	}
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
	err = db.MySQL.Model(&model.Video{}).Create(&video).Error
	if err != nil {
		return err
	}
	return nil
}
