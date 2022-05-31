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

func Separate(){
	
}