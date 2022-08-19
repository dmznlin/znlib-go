package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	str, err := RedisClient.Ping()
	if str == "PONG" {
		t.Log(str)
	} else {
		t.Error(err)
	}
}

func TestLock(t *testing.T) {
	rg := NewRoutineGroup()
	tag := "znlib.serialid"
	lock := RedisClient.Lock(tag, 3*time.Second, 50*time.Second)

	rg.Run(func(arg ...interface{}) {
		str, err := RedisClient.Get(Application.Ctx, tag).Result()
		if err == nil {
			Info(str)
		} else {
			Error(err)
		}
	})

	rg.Run(func(arg ...interface{}) {
		for i := 0; i < 5; i++ {
			str, err := SerialID.DateID("test", 9)
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
