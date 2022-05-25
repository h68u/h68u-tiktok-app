package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/minio/minio-go/v6"
)

var Client *minio.Client

var AliyunClient *oss.Client

func Init() {
	MinioInit()
	AliyunInit()
}
