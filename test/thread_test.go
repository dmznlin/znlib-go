package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRoutineGroupRun(t *testing.T) {
	var count int32
	group := NewRoutineGroup()
	for i := 0; i < 3; i++ {
		group.Run(func(arg ...interface{}) {
			atomic.AddInt32(&count, 1)
		})
	}

	group.Wait()
	if count != 3 {
		t.Errorf("znlib.RoutineGroupRun wrong")
	}
}

func TestRoutingGroupRunSafe(t *testing.T) {
	var count int32
	group := NewRoutineGroup()
	var once sync.Once
	for i := 0; i < 3; i++ {
		group.RunSafe(func(arg ...interface{}) {
			once.Do(func() {
				panic("hello")
			})
			atomic.AddInt32(&count, 1)
		})
	}

	group.Wait()
	if count != 2 {
		t.Errorf("znlib.RoutingGroupRunSafe wrong")
	}
}

func TestRoutineWidthParams(t *testing.T) {
	group := NewRoutineGroup()
	group.Run(func(arg ...interface{}) {
		v, _ := arg[2].(int) //第二个参数
		if v != 3 {
			t.Errorf("znlib.RoutingGroupRunSafe wrong")
		}
	}, 1, 2, 3)
}
