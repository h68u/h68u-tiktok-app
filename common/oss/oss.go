package oss

import "github.com/minio/minio-go/v6"

var Client *minio.Client

func Init() {
	MinioInit()
}
