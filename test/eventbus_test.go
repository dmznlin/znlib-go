package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestEventBus(t *testing.T) {
	fn := func(str string) {
		t.Log(str)
	}
	fn1 := func(str string, val int) {
		t.Logf("str: %s val:%d", str, val)
	}
	event := NewEventBus()
	event.SubscribeOnce("a", fn)
	event.SubscribeAsync("b", fn1)
	event.Subscribe("a", fn)
	event.Subscribe("b", fn1)
	//订阅主题

	event.Publish("a", "a,fn")
	event.Publish("b", "b.fn1", 222)
	//发布主题数据

	event.Unsubscribe("a", fn1)
	event.WaitAsync()
	//退订和等待
}
