package test

import (
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestContains(t *testing.T) {
	if znlib.Contains(3, []int{1, 2, 3, 4}) != 2 {
		t.Errorf("znlib.Contains error")
	}
}

func TestIsNil(t *testing.T) {
	var app []znlib.LogConfig
	//app = make([]znlib.LogConfig, 0)
	if znlib.IsNil(app) != true {
		t.Error("znlib.Isnil error")
	}
}

func TestIsNumber(t *testing.T) {
	_, ok := znlib.IsNumber("12.3", false)
	if ok != false {
		t.Errorf("znlib.Isnumber error")
	}
}
