package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestCopy(t *testing.T) {
	if Copy("hello", 0, 2) != "he" ||
		Copy("hello", 4, 1) != "o" {
		t.Errorf("znlib.Copy wrong")
	}
}

func TestTrim(t *testing.T) {
	if Trim("  str ing  \n") != "str ing" {
		t.Errorf("znlib.Trim wrong")
	}
}

func TestCopyLeft(t *testing.T) {
	if CopyLeft("!!hello", 2) != "!!" {
		t.Errorf("znlib.CopyLeft wrong")
	}
}

func TestCopyRight(t *testing.T) {
	if CopyRight("hello!!", 2) != "!!" ||
		CopyRight("!hello!", 20) != "!hello!" {
		t.Errorf("znlib.CopyRight wrong")
	}
}

func TestStrPos(t *testing.T) {
	if StrPos("Hello,Word", "Lo,w") != 3 {
		t.Errorf("znlib.StrPos wrong")
	}
}

func TestStrReplace(t *testing.T) {
	if StrReplace("中文English", "0", "文", "eng") != "中00lish" {
		t.Errorf("znlib.StrReplace wrong")
	}
}

func TestStr2Bit(t *testing.T) {
	if Str2Bit("00000101") != 5 {
		t.Errorf("znlib.Str2Bit wrong")
	}
}

func TestBit2Str(t *testing.T) {
	if Bit2Str(5) != "00000101" {
		t.Errorf("znlib.Bit2Str wrong")
	}
}

func TestStrReverse(t *testing.T) {
	if StrReverse("12345") != "54321" {
		t.Errorf("znlib.StrReverse wrong")
	}
}
