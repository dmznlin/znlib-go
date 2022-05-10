package znlib

import (
	"strings"
)

/******************************************************************************
作者: dmzn@163.com 2022-05-09
描述: 常用字符串相关函数
******************************************************************************/

//Date: 2022-05-09
//Parm: 字符串;开始位置;长度
//Desc: 从start开始,复制len长的str子字符串
func Copy(str string, start, length int) string {
	maxLen := len(str)
	if start > maxLen-1 { //超出索引
		return ""
	}

	length = start + length
	if length > maxLen {
		length = maxLen
	}
	return str[start:length]
}

//Date: 2022-05-09
//Parm: 字符串
//Desc: 清除两端的空格、回车换行符
func Trim(str string) string {
	return strings.Trim(str, string([]byte{10, 13, 32}))
}
