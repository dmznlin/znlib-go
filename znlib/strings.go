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

// Key define
type Key = uint16

// Key constants
const (
	KeyF1 Key = 0xFFFF - iota
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPgup
	KeyPgdn
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	key_min // see terminfo
	MouseLeft
	MouseMiddle
	MouseRight
	MouseRelease
	MouseWheelUp
	MouseWheelDown
)

const (
	KeyCtrlTilde      Key = 0x00
	KeyCtrl2          Key = 0x00
	KeyCtrlSpace      Key = 0x00
	KeyCtrlA          Key = 0x01
	KeyCtrlB          Key = 0x02
	KeyCtrlC          Key = 0x03
	KeyCtrlD          Key = 0x04
	KeyCtrlE          Key = 0x05
	KeyCtrlF          Key = 0x06
	KeyCtrlG          Key = 0x07
	KeyBackspace      Key = 0x08
	KeyCtrlH          Key = 0x08
	KeyTab            Key = 0x09
	KeyCtrlI          Key = 0x09
	KeyCtrlJ          Key = 0x0A
	KeyCtrlK          Key = 0x0B
	KeyCtrlL          Key = 0x0C
	KeyEnter          Key = 0x0D
	KeyCtrlM          Key = 0x0D
	KeyCtrlN          Key = 0x0E
	KeyCtrlO          Key = 0x0F
	KeyCtrlP          Key = 0x10
	KeyCtrlQ          Key = 0x11
	KeyCtrlR          Key = 0x12
	KeyCtrlS          Key = 0x13
	KeyCtrlT          Key = 0x14
	KeyCtrlU          Key = 0x15
	KeyCtrlV          Key = 0x16
	KeyCtrlW          Key = 0x17
	KeyCtrlX          Key = 0x18
	KeyCtrlY          Key = 0x19
	KeyCtrlZ          Key = 0x1A
	KeyEsc            Key = 0x1B
	KeyCtrlLsqBracket Key = 0x1B
	KeyCtrl3          Key = 0x1B
	KeyCtrl4          Key = 0x1C
	KeyCtrlBackslash  Key = 0x1C
	KeyCtrl5          Key = 0x1D
	KeyCtrlRsqBracket Key = 0x1D
	KeyCtrl6          Key = 0x1E
	KeyCtrl7          Key = 0x1F
	KeyCtrlSlash      Key = 0x1F
	KeyCtrlUnderscore Key = 0x1F
	KeySpace          Key = 0x20
	KeyBackspace2     Key = 0x7F
	KeyCtrl8          Key = 0x7F
)

const (
	LayoutTime          = "15:04:05"                //时间
	LayoutTimeMilli     = "15:04:05.000"            //时间 + 毫秒
	LayoutDate          = "2006-01-02"              //日期
	LayoutDateTime      = "2006-01-02 15:04:05"     //日期 + 时间
	LayoutDateTimeMilli = "2006-01-02 15:04:05.000" //日期 + 时间 + 毫秒
)

//--------------------------------------------------------------------------------

/*StrCopy 2022-05-30 13:40:27
  参数: str,字符串
  参数: start,开始位置
  参数: length,长度
  描述: 从start开始,复制len长的str子字符串
*/
func StrCopy(str string, start, length int) string {
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

/*StrCopyLeft 2022-05-30 13:41:02
  参数: str,字符串
  参数: length,长度
  描述: 从str开始位置复制长度为length的子字符串
*/
func StrCopyLeft(str string, length int) string {
	maxLen := len(str)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return str[:length]
}

/*StrCopyRight 2022-05-30 13:41:20
  参数: str,字符串
  参数: length,长度
  描述: 从str末尾向前复制长度为length的子字符串
*/
func StrCopyRight(str string, length int) string {
	maxLen := len(str)
	if maxLen < 1 {
		return ""
	}

	if length > maxLen {
		length = maxLen
	}
	return str[maxLen-length:]
}

/*StrTrim 2022-05-30 13:41:48
  参数: str,字符串
  描述: 清除两端的制表、空格、回车换行符
*/
func StrTrim(str string) string {
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
	defer ErrorHandle(false, func(err any) {
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
	defer ErrorHandle(false, func(err any) {
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
