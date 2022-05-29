package util

import (
	"strings"
)

func connect(str1 int64,str2 int64)(string){
	var str strings.Builder
	str.WriteString(string(str1))
	str.WriteString("::")
	str.WriteString(string(str2))
	return str.String()

}

func separatesplite