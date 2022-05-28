package oss

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/policy"
	"go.uber.org/zap"
	"io"
	"net/url"
	"tikapp/common/config"
	"tikapp/common/log"
	"time"
)

func MinioInit() {
	var err error
	Client, err = minio.New(config.MinioCfg.Endpoint, config.MinioCfg.AccessKeyID, config.MinioCfg.SecretAccessKey, false)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// CreateMinoBucket 创建minio 桶
func CreateMinoBucket(bucketName string) {
	location := "us-east-1"
	err := Client.MakeBucket(bucketName, location)
	if err != nil {
		// 检查存储桶是否已经存在。
		exists, err := Client.BucketExists(bucketName)
		if err == nil && exists {
			log.Logger.Error(fmt.Sprintf("We already own %s\n", bucketName))
		} else {
			log.Logger.Error("create bucket error")
			return
		}
	}
	// 设置存储桶访问权限
	err = Client.SetBucketPolicy(bucketName, policy.BucketPolicyReadWrite)

	if err != nil {
		log.Logger.Error("set bucket policy error")
		return
	}
	log.Logger.Info(fmt.Sprintf("Successfully created %s\n", bucketName))
}

// UploadVideo 上传文件给minio指定的桶中
func UploadVideo(bucketName, objectName string, reader io.Reader, objectSize int64) (ok bool) {
	n, err := Client.PutObject(bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		log.Logger.Error("uploadFile error")
		return false
	}
	log.Logger.Info(fmt.Sprintf("Successfully uploaded bytes: %v", n))
	return true
}

// GetVideoUrl 获取文件url
func GetVideoUrl(bucketName string, fileName string, expires time.Duration) string {
	//time.Second*24*60*60
	reqParams := make(url.Values)
	presignedURL, err := Client.PresignedGetObject(bucketName, fileName, expires, reqParams)
	if err != nil {
		zap.L().Error(err.Error())
		return ""
	}
	return presignedURL.String()
}
