// Package pinyin
/******************************************************************************
  作者: dmzn@163.com 2024-07-03 13:06:50
  描述: 拼音有关的函数
******************************************************************************/
package pinyin

import "strings"

// Str2Pinyin 2024-01-03 15:06:07
/*
 参数: str,字符串
 参数: named,人名模式
 描述: 获取 str 的拼音首字母
*/
func Str2Pinyin(str string, named ...bool) string {
	var buf string
	if named != nil && named[0] {
		buf = NewDict().Name(str, ",").None()
	} else {
		buf = NewDict().Convert(str, ",").None()
	}

	words := strings.Split(buf, ",")
	res := make([]byte, 0, len(words))
	for _, v := range words {
		if len(v) > 0 {
			res = append(res, v[0])
		}
	}

	return string(res)
}
