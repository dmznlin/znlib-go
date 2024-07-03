package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/dmznlin/znlib-go/znlib/bin"
	"github.com/dmznlin/znlib-go/znlib/pinyin"
	"testing"
	"time"
)

func TestCopy(t *testing.T) {
	if StrCopy("hello", 0, 2) != "he" ||
		StrCopy("hello", 4, 1) != "o" {
		t.Errorf("znlib.Copy wrong")
	}
}

func TestTrim(t *testing.T) {
	if StrTrim("  str ing  \n") != "str ing" {
		t.Errorf("znlib.Trim wrong")
	}
}

func TestCopyLeft(t *testing.T) {
	if StrCopyLeft("!!hello", 2) != "!!" {
		t.Errorf("znlib.CopyLeft wrong")
	}
}

func TestCopyRight(t *testing.T) {
	if StrCopyRight("hello!!", 2) != "!!" ||
		StrCopyRight("!hello!", 20) != "!hello!" {
		t.Errorf("znlib.CopyRight wrong")
	}
}

func TestStrPos(t *testing.T) {
	if StrPos("Hello,Word", "Lo,w") != 3 {
		t.Errorf("znlib.StrPos wrong")
	}
}

func TestStrIn(t *testing.T) {
	if !StrIn("a", "b", "a", "c") {
		t.Errorf("znlib.StrIn wrong")
	}
}

func TestStrIndex(t *testing.T) {
	if StrIndex(2, "a", "b", "c") != "c" {
		t.Errorf("znlib.StrIndex wrong")
	}
}

func TestStrReplace(t *testing.T) {
	if StrReplace("中文English", "0", "文", "eng") != "中00lish" {
		t.Errorf("znlib.StrReplace wrong")
	}
}

func TestStr2Bit(t *testing.T) {
	if bin.Str2Bit("00000101") != 5 {
		t.Errorf("znlib.Str2Bit wrong")
	}
}

func TestBit2Str(t *testing.T) {
	if bin.Bit2Str(5) != "00000101" {
		t.Errorf("znlib.Bit2Str wrong")
	}
}

func TestStrReverse(t *testing.T) {
	if StrReverse("12345") != "54321" {
		t.Errorf("znlib.StrReverse wrong")
	}
}

func TestStr2Date(t *testing.T) {
	lc := time.Local
	if Str2DateTime("2022-06-01 12:00:01") != time.Date(2022, 06, 01, 12, 00, 01, 0, lc) {
		t.Errorf("znlib.Str2DateTime wrong")
	}
}

func TestDate2Str(t *testing.T) {
	if DateTime2Str(time.Date(2022, 06, 01, 00, 00, 00, 00, time.Local), LayoutDate) != "2022-06-01" {
		t.Errorf("znlib.DateTime2Str wrong")
	}
}

func TestStrDel(t *testing.T) {
	var str = "中eng文混杂字符串"
	if StrDel(str, 0, 9) != "" {
		t.Errorf("znlib.Strdel wrong")
	}
}

func TestStr2Pinyin(t *testing.T) {
	var str = "白日依山尽"
	if pinyin.Str2Pinyin(str) != "brysj" {
		t.Errorf("znlib.Str2Pinyin wrong")
	}
}

func TestStrHex2Bin(t *testing.T) {
	var hex = []byte("4a 4b 4c 4d5 a")
	bin, ok := StrHex2Bin(hex, true)
	if !ok || string(bin) != "JKLMZ" {
		t.Errorf("znlib.StrHex2Bin wrong")
	}
}

func TestStrBin2Hex(t *testing.T) {
	var bin = []byte("JKLMZ")
	hex, ok := StrBin2Hex(bin, false)
	if !ok || string(hex) != "4a4b4c4d5a" {
		t.Errorf("znlib.StrHex2Bin wrong")
	}

	hex, ok = StrBin2Hex(bin, true)
	if !ok || string(hex) != "4a 4b 4c 4d 5a" {
		t.Errorf("znlib.StrHex2Bin.blank wrong")
	}

	hex, ok = StrBin2Hex([]byte("中文"), false)
	if ok {
		t.Log(string(hex))
	}
}
