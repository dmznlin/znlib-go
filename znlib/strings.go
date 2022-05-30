/*Package znlib ***************************************************************
作者: dmzn@163.com 2022-05-30 13:40:08
描述: 常用字符串相关函数
******************************************************************************/
package znlib

import (
	"github.com/dmznlin/znlib-go/znlib/biu"
	"strings"
	"unicode"
)

/*Copy 2022-05-30 13:40:27
参数: str,字符串
参数: start,开始位置
参数: length,长度
描述: 从start开始,复制len长的str子字符串
*/
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

/*CopyLeft 2022-05-30 13:41:02
  参数: str,字符串
  参数: length,长度
  描述: 从str开始位置复制长度为length的子字符串
*/
func CopyLeft(str string, length int) string {
	maxLen := len(str)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return str[:length]
}

/*CopyRight 2022-05-30 13:41:20
  参数: str,字符串
  参数: length,长度
  描述: 从str末尾向前复制长度为length的子字符串
*/
func CopyRight(str string, length int) string {
	maxLen := len(str)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return str[maxLen-length:]
}

/*Trim 2022-05-30 13:41:48
  参数: str,字符串
  描述: 清除两端的制表、空格、回车换行符
*/
func Trim(str string) string {
	return strings.Trim(str, string([]byte{9, 10, 13, 32}))
}

/*StrReplace 2022-05-30 13:43:26
  参数: str,字符串
  参数: new,新字符串
  参数: old,现有字符串
  描述: 使用new替换str中的old字符串,不区分大小写
*/
func StrReplace(str string, new string, old ...string) string {
	if old == nil || len(str) < 1 {
		return str
	}

	var idx, pos, sublen int
	var update bool = true //需更新strBuf
	var strBuf = make([]rune, 0, 20)
	var subBuf = make([]rune, 0, 10)

	for _, tmp := range old {
		subBuf = append(subBuf[0:0], []rune(tmp)...)
		sublen = len(subBuf)
		if sublen < 1 { //旧字符串为空
			continue
		}

		if update {
			update = false
			strBuf = append(strBuf[0:0], []rune(str)...)
		}

		idx = 0
		pos = StrPosFrom(strBuf, subBuf, idx)
		for pos >= 0 {
			update = true
			if idx == 0 {
				str = ""
				//重新配置字符串
			}

			str = str + string(strBuf[idx:pos]) + new
			idx = pos + sublen
			pos = StrPosFrom(strBuf, subBuf, idx)
		}

		if update {
			if idx < len(strBuf) {
				str = str + string(strBuf[idx:])
			}

		}
	}

	return str
}

/*StrPosFrom 2022-05-30 13:44:04
  参数: str,字符串
  参数: sub,子字符串
  参数: from,开始索引
  描述: 检索sub在str中的位置,不区分大小写
*/
func StrPosFrom(str, sub []rune, from int) int {
	lstr := len(str)
	lsub := len(sub)
	if lstr < 1 || lsub < 1 {
		return -1
	}

	compare := func(a, b rune) bool {
		return a == b || (unicode.IsLower(a) && unicode.ToUpper(a) == b) || (unicode.IsUpper(a) && unicode.ToLower(a) == b)
		//忽略大小写
	}

	var match bool
	for idx := from; idx < lstr; idx++ {
		if !compare(str[idx], sub[0]) {
			continue
			//匹配首字母
		}

		match = true
		for i := 1; i < lsub; i++ {
			if idx+i >= lstr {
				match = false
				break
				//已超出字符串长度
			}

			if !compare(str[idx+i], sub[i]) {
				match = false
				break
				//子字符串未匹配
			}
		}

		if match {
			return idx
		}
	}

	return -1
}

/*StrPos 2022-05-30 13:44:34
  参数: str,字符串
  参数: sub,子字符串
  描述: 检索sub在str中的位置,不区分大小写
*/
func StrPos(str string, sub string) int {
	return StrPosFrom([]rune(str), []rune(sub), 0)
}

/*Str2Bit 2022-05-30 13:27:53
  参数: str,字符串
  描述: 使用str构建一个字节
*/
func Str2Bit(str string) (ret byte) {
	defer ErrorHandle(false, func(err any) {
		if err != nil {
			ret = 0
		}
	})

	return biu.BinaryStringToBytes(str)[0]
}

/*Bit2Str 2022-05-30 14:25:51
  参数: val,字节
  描述: 返回val的二进制描述
*/
func Bit2Str(val byte) string {
	return biu.ByteToBinaryString(val)
}
