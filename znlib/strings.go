/*Package znlib ***************************************************************
作者: dmzn@163.com 2022-05-30 13:40:08
描述: 常用字符串相关函数
******************************************************************************/
package znlib

import (
	"github.com/dmznlin/znlib-go/znlib/biu"
	"strings"
	"time"
	"unicode"
)

/*StrCopy 2022-05-30 13:40:27
  参数: str,字符串
  参数: start,开始位置
  参数: length,长度
  描述: 从start开始,复制len长的str子字符串
*/
func StrCopy(str string, start, length int) string {
	buf := []rune(str)
	maxLen := len(buf)
	if maxLen < 1 || start >= maxLen { //超出索引
		return ""
	}

	length = start + length
	if length > maxLen {
		length = maxLen
	}
	return string(buf[start:length])
}

/*StrCopyLeft 2022-05-30 13:41:02
  参数: str,字符串
  参数: length,长度
  描述: 从str开始位置复制长度为length的子字符串
*/
func StrCopyLeft(str string, length int) string {
	buf := []rune(str)
	maxLen := len(buf)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return string(buf[:length])
}

/*StrCopyRight 2022-05-30 13:41:20
  参数: str,字符串
  参数: length,长度
  描述: 从str末尾向前复制长度为length的子字符串
*/
func StrCopyRight(str string, length int) string {
	buf := []rune(str)
	maxLen := len(buf)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return string(buf[maxLen-length:])
}

/*StrTrim 2022-05-30 13:41:48
  参数: str,字符串
  描述: 清除两端的制表、空格、回车换行符
*/
func StrTrim(str string) string {
	return strings.Trim(str, string([]byte{9, 10, 13, 32}))
}

/*StrDel 2022-06-17 16:05:54
  参数: str,字符串
  参数: from,开始位置
  参数: end,结束位置
  描述: 从str中删除from-end子字符串
*/
func StrDel(str string, from, end int) string {
	buf := []rune(str)
	maxLen := len(buf)
	if maxLen < 1 || end < from {
		return ""
	}

	if from < 0 {
		from = 0
	}
	if end >= maxLen-1 {
		return string(buf[0:from])
	} else {
		return string(append(buf[0:from], buf[end+1:]...))
	}
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
	var update = true //需更新strBuf
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

/*StrIn 2022-07-19 20:00:36
  参数: str,字符串
  参数: array,字符串数组
  描述: 判断str是否在array中,不区分大小写
*/
func StrIn(str string, array ...string) bool {
	for _, in := range array {
		if strings.EqualFold(in, str) {
			return true
		}
	}

	return false
}

/*Str2Bit 2022-05-30 13:27:53
  参数: str,字符串
  描述: 使用str构建一个字节
*/
func Str2Bit(str string) (ret byte) {
	defer DeferHandle(false, "znlib.Str2Bit", func(err any) {
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

/*StrReverse 2022-05-30 21:46:04
  参数: str,字符串
  描述: 将str首尾翻转
*/
func StrReverse(str string) string {
	runes := []rune(str)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

//--------------------------------------------------------------------------------

/*DateTime2Str 2022-06-01 13:57:37
  参数: dt,时间值
  参数: fmt,格式
  描述: 使用fmt格式转换dt为字符串
*/
func DateTime2Str(dt time.Time, fmt ...string) (ret string) {
	defer DeferHandle(false, "znlib.DateTime2Str", func(err any) {
		if err != nil {
			ret = time.Now().Format(LayoutDateTime)
		}
	})

	var lay string
	if fmt == nil {
		nD := dt.Year() > 0
		nT := !nD || dt.Hour() > 0 || dt.Minute() > 0 || dt.Second() > 0 || dt.Nanosecond() > 0

		switch {
		case nD && nT:
			lay = LayoutDateTime
		case nD:
			lay = LayoutDate
		case nT:
			lay = LayoutTime
		default:
			lay = LayoutDateTime
		}
	} else {
		lay = fmt[0]
	}

	return dt.Format(lay)
}

/*Str2DateTime 2022-06-01 14:06:49
  参数: dt,时间字符串
  参数: fmt,格式
  描述: 使用fmt格式转换dt为时间值
*/
func Str2DateTime(dt string, fmt ...string) (ret time.Time) {
	defer DeferHandle(false, "znlib.Str2DateTime", func(err any) {
		if err != nil {
			ret = time.Now()
		}
	})

	if len(dt) < 1 {
		panic("znlib.Str2DateTime: dt is empty")
	}

	var lay string
	if fmt == nil { //default layout
		id := strings.Index(dt, "-") > 0
		it := strings.Index(dt, ":") > 0

		switch {
		case id && it:
			lay = LayoutDateTime
		case id:
			lay = LayoutDate
		case it:
			lay = LayoutTime
		default:
			lay = LayoutDateTime
		}

	} else {
		lay = fmt[0]
	}

	ret, err := time.ParseInLocation(lay, dt, time.Local)
	if err != nil {
		panic(err)
	}
	return
}
