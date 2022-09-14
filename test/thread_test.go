package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

func TestRoutineInRange(t *testing.T) {
	array := []string{"a", "b", "c"}
	group := NewRoutineGroup()
	for _, v := range array {
		group.Run(func(arg ...interface{}) {
			t.Logf("first: %s", v)
			//routine启动后,v值保持在最后一个
		})
	}

	for _, v := range array {
		group.Run(func(arg ...interface{}) {
			t.Logf("second: %s", arg[0])
		}, v) //参数复制v值
	}
	group.Wait()
}

func testRoutineInRange_fun(t *testing.T, wg *sync.WaitGroup, str string) {
	t.Logf("in routine: %s", str)
	wg.Done()
}
func TestRoutineInRange2(t *testing.T) {
	wg := &sync.WaitGroup{}
	array := []string{"a", "b", "c"}
	for _, v := range array {
		wg.Add(1)
		go testRoutineInRange_fun(t, wg, v)
	}

	wg.Wait()
}

func TestRunWait(t *testing.T) {
	group := NewRoutineGroup()
	err := group.WaitRun(2*time.Second, func(args ...interface{}) {
		time.Sleep(1 * time.Second)
		//panic(args[0])
	}, "Hello Panic")

	if err == nil {
		t.Logf("znlib.WaitRun no timeout.")
	} else {
		t.Errorf(err.Error())
	}
}
