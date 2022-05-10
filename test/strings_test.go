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
