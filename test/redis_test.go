package test

import (
	"testing"
	"time"

	. "github.com/dmznlin/znlib-go/znlib"
	. "github.com/dmznlin/znlib-go/znlib/redis"
)

func TestPing(t *testing.T) {
	str, err := Client.Ping()
	if str == "PONG" {
		t.Log(str)
	} else {
		t.Error(err)
	}
}

func TestLock(t *testing.T) {
	rg := NewRoutineGroup()
	tag := "znlib.serialid"
	lock := Client.Lock(tag, 3*time.Second, 50*time.Second)

	rg.Run(func(arg ...interface{}) {
		str, err := Client.Get(Application.Ctx, tag).Result()
		if err == nil {
			Info(str)
		} else {
			Error(err)
		}
	})

	rg.Run(func(arg ...interface{}) {
		for i := 0; i < 5; i++ {
			str, err := Client.DateID("test", 9)
			if err == nil {
				Info(str)
			} else {
				Error(err)
			}
		}

	})

	rg.Wait()
	lock.Unlock()
}
