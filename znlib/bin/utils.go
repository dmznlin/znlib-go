package bin

import (
	"github.com/dmznlin/znlib-go/znlib"
	"github.com/dmznlin/znlib-go/znlib/bin/biu"
)

// Str2Bit 2022-05-30 13:27:53
/*
 参数: str,字符串
 描述: 使用str构建一个字节
*/
func Str2Bit(str string) (ret byte) {
	defer znlib.DeferHandle(false, "znlib.strings.Str2Bit", func(err error) {
		if err != nil {
			ret = 0
		}
	})

	return biu.BinaryStringToBytes(str)[0]
}

// Bit2Str 2022-05-30 14:25:51
/*
 参数: val,字节
 描述: 返回val的二进制描述
*/
func Bit2Str(val byte) string {
	return biu.ByteToBinaryString(val)
}
