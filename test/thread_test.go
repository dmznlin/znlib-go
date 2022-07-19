package test

import (
	. "github.com/dmznlin/znlib-go/znlib/threading"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRoutineGroupRun(t *testing.T) {
	var count int32
	group := NewRoutineGroup()
	for i := 0; i < 3; i++ {
		group.Run(func() {
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
		group.RunSafe(func() {
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
