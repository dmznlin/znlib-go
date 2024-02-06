package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
	"time"
)

// znlib初始化
var _ = InitLib(func() {
	Application.SetWorkDir(`D:\Mywork\KTManager\znlib-go\main\bin\`)
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

func TestWaitFor(t *testing.T) {
	dura := 3 * time.Second
	end := time.Now().Add(dura)
	WaitFor(dura, func() bool {
		Info("in func.")
		return false
	})

	if time.Now().Before(end) {
		t.Error("Waitfor exit.")
	}
}

func TestDeferHander(t *testing.T) {
	defer DeferHandle(false, "test", func(err error) {
		t.Log(err)
	})

	panic(ErrorMsg(nil, "hello,error"))
}
