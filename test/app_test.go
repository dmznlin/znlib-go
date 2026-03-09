package test

import (
	"testing"
	"time"

	. "github.com/dmznlin/znlib-go/znlib"
)

// znlib初始化
var _ = InitLib(func() {
	Application.SetWorkDir(`D:\Program Files\MyLib\znlib-go\main\bin\`)

	cfg := struct {
		Age  int    `json:"age"`
		Addr string `json:"addr"`
	}{
		Age:  10,
		Addr: "127.0.0.1:8080",
	}

	GlobalConfig.App = &cfg
	//添加外部配置
}, func() {
	Info(GlobalConfig.App)
	//读取外部配置
})

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
	dura := 50 * time.Millisecond
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
