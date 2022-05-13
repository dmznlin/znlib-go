package test

import (
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestCopy(t *testing.T) {
	if znlib.Copy("hello", 0, 2) != "he" ||
		znlib.Copy("hello", 4, 1) != "o" {
		t.Errorf("znlib.Copy wrong")
	}
}

func TestTrim(t *testing.T) {
	if znlib.Trim("  str ing  \n") != "str ing" {
		t.Errorf("znlib.Trim wrong")
	}
}

func TestCopyLeft(t *testing.T) {
	if znlib.CopyLeft("!!hello", 2) != "!!" {
		t.Errorf("znlib.CopyLeft wrong")
	}
}

func TestCopyRight(t *testing.T) {
	if znlib.CopyRight("hello!!", 2) != "!!" ||
		znlib.CopyRight("!hello!", 20) != "!hello!" {
		t.Errorf("znlib.CopyRight wrong")
	}
}

func TestStrPos(t *testing.T) {
	if znlib.StrPos("Hello,Word", "Lo,w") != 3 {
		t.Errorf("znlib.StrPos wrong")
	}
}

func TestStrReplace(t *testing.T) {
	if znlib.StrReplace("中文English", "0", "文", "eng") != "中00lish" {
		t.Errorf("znlib.StrReplace wrong")
	}
}
