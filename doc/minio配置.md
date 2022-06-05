# 1.环境搭建

本地和用线上服务器两个方案
## 1.1 docker 搭建 minio 测试环境
- **创建一个 docker-compose.yaml , 填入以下内容**
```
version: "3.5"

services:
   mingyue-minio:
    image: minio/minio:RELEASE.2021-06-17T00-10-46Z
    container_name: my-minio
    restart: always
    command: server /data
    ports:
      - 9000:9000
    volumes:
      - ./minio/data:/data
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=admin12345678
```

- **执行 `docker-compose up -d`**
- **修改配置文件配置**

```
minio:
  endpoint: 127.0.0.1:9000
  accessKeyID: admin
  secretAccessKey: admin12345678
```


## 1.2 直接用线上服务器 
- **配置文件配置**
```
minio:
  endpoint: 
  accessKeyID: 
  secretAccessKey: 
```

# 2. [minio-go API文档](http://docs.minio.org.cn/docs/master/golang-client-api-reference#PutObject)

# 3. 使用样例
###  3.1 文件上传和url获取
```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"tikapp/common/oss"
	"time"
)

const BucketName = "test"

func main() {
	// 创建一个桶,相当于一个存储的容器
	oss.CreateMinoBuket(BucketName)
	// 打开要上传的文件
	file, _ := os.Open("good.mp4")
	reader := bufio.NewReader(file)
	//上传文件, -1 代表文件大小未知
	ok := oss.UploadVideo(BucketName, file.Name(), reader, -1)
	if !ok {
		fmt.Println("upload file error")
	}
	// 获取已上传文件的url, 设置了有效期限
	fileUrl := oss.GetVideoUrl(BucketName, "good.mp4", time.Second*24*60*60)
	fmt.Println(fileUrl)
}
```
