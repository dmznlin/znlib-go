package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
)

// znlib初始化
var _ = InitLib(func() {
	AppPath = `D:\Program Files\MyVCL\go\znlib-go\main\bin`
	// Application.ConfigFile = `c:\main.ini`
}, nil)

func Test_FixPath(t *testing.T) {
	last := Application.ExePath[len(Application.ExePath)-1]
	if Application.IsWindows && last != '\\' {
		t.Errorf("Test_FixPath wrong")
	}
}

func TestTryFinally(t *testing.T) {
	err := TryFinal{Try: func() error {
		panic("raise 1")
	}, Finally: func() {
		t.Log("i am finally")
	}, Except: func(err error) {

	}}.Run()

	if err == nil {
		t.Errorf("znlib.TryFinal wrong")
	}
}
