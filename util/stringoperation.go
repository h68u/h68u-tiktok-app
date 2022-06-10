package util

import (
	"strings"
	"strconv"
)

func Connect(str1 int64,str2 int64)(string){
	var str strings.Builder
	str.WriteString(strconv.FormatInt(str1,10))
	str.WriteString("::")
	str.WriteString(strconv.FormatInt(str2,10))
	return str.String()

}

func Separate(Id string)(int64,int64){
	StringId := strings.Split(Id,"::")
	videoId := StringId[0]
	userId := StringId[1]
	video, _ := strconv.ParseInt(videoId,10,64)
	user, _ := strconv.ParseInt(userId,10,64)
	return video,user
	
}