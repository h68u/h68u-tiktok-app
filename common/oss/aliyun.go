package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"mime/multipart"
	"tikapp/common/config"
)

func AliyunInit() {
	client, err := oss.New(config.AliyunCfg.Endpoint, config.AliyunCfg.AccessKeyID, config.AliyunCfg.AccessKeySecret)
	if err != nil {
		fmt.Println(err)
		return
	}
	AliyunClient = client
}

func CreateBucket(name string) {
	err := AliyunClient.CreateBucket(name, oss.ACL(oss.ACLPublicReadWrite))
	if err != nil {
		exist, err := AliyunClient.IsBucketExist(name)
		if err == nil && exist {
			logger.Infof("We already own %s\n", name)
		} else {
			logger.Error("create bucket error")
			return
		}
	}
	logger.Infof("Successfully created %s\n", name)
}

func UploadVideoToOss(bucketName string, objectName string, reader multipart.File) (bool, error) {
	bucket, err := AliyunClient.Bucket(bucketName)
	if err != nil {
		return false, err
	}
	err = bucket.PutObject(objectName, reader)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func GetOssVideoUrlAndImgUrl(bucketName string, objectName string) (string, string, error) {
	url := "https://" + bucketName + "." + config.AliyunCfg.Endpoint + "/" + objectName
	return url, url + "?x-oss-process=video/snapshot,t_0,f_jpg,w_0,h_0,m_fast,ar_auto", nil
}
