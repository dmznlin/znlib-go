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
	ok := TryFinally(func() {
		panic("raise exception")
	},
		func() {
			t.Log("i am finally")
		},

		func(err any) {
			t.Log("again")
		},
	)

	if ok {
		t.Errorf("znlib.TryFinally wrong")
	}
}
