package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func Test_FixPath(t *testing.T) {
	last := Application.ExePath[len(Application.ExePath)-1]
	if Application.IsWindows && last != '\\' {
		t.Errorf("Test_FixPath wrong")
	}
}

func TestTryFinally(t *testing.T) {
	ok := TryFinal{Try: func() {
		panic("raise 1")
	}, Finally: func() {
		t.Log("i am finally")
	}}.Run()

	if ok {
		t.Errorf("znlib.TryFinal wrong")
	}
}
