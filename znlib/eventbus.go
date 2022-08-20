/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-08-19 09:25:21
  描述: 事件总线

  参考: https://github.com/asaskevich/EventBus.git
  作者: Alex Saskevich
  协议: The MIT License (MIT)
******************************************************************************/
package znlib

import (
	"fmt"
	"reflect"
	"sync"
)

//EventBus 总线实现
type EventBus struct {
	handlers map[string][]*eventHandler
	lock     sync.RWMutex
	rg       *RoutineGroup
}

//eventHandler 事件句柄
type eventHandler struct {
	callBack reflect.Value //回调函数
	flagOnce bool          //单次调用标识
	async    bool          //异步调用标识
}

/*NewEventBus 2022-08-19 16:49:26
  描述: 生成事件总线
*/
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]*eventHandler),
		rg:       NewRoutineGroup(),
	}
}

/*doSubscribe 2022-08-19 13:39:26
  参数: topic,主题
  参数: fn,函数
  参数: handler,事件句柄
  描述: 为topic注册fn处理函数
*/
func (bus *EventBus) doSubscribe(topic string, fn interface{}, handler *eventHandler) error {
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("znlib.EventBus: subscribe.fn must be func")
	}

	bus.lock.Lock()
	defer bus.lock.Unlock()
	//lock first

	bus.handlers[topic] = append(bus.handlers[topic], handler)
	return nil
}

func (bus *EventBus) Subscribe(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn,
		&eventHandler{
			callBack: reflect.ValueOf(fn),
		})
}

func (bus *EventBus) SubscribeAsync(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn,
		&eventHandler{
			callBack: reflect.ValueOf(fn),
			async:    true,
		})
}

func (bus *EventBus) SubscribeOnce(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn,
		&eventHandler{
			callBack: reflect.ValueOf(fn),
			flagOnce: true,
		})
}

func (bus *EventBus) SubscribeOnceAsync(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn,
		&eventHandler{
			callBack: reflect.ValueOf(fn),
			flagOnce: true,
			async:    true,
		})
}

func (bus *EventBus) HasCallback(topic string) bool {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	if _, ok := bus.handlers[topic]; ok {
		return len(bus.handlers[topic]) > 0
	} else {
		return false
	}
}

/*Unsubscribe 2022-08-20 18:53:04
  参数: topic,主题
  参数: fn,函数
  描述: 删除topic主题中的fn函数
*/
func (bus *EventBus) Unsubscribe(topic string, fn ...interface{}) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if _, ok := bus.handlers[topic]; ok {
		if fn == nil { //delete topic
			delete(bus.handlers, topic)
			return nil
		}

		return bus.deleteHandler(topic, nil, fn...)
	} else {
		return fmt.Errorf("znlib.Unsubscribe: topic %s doesn't exist", topic)
	}
}

/*Publish 2022-08-19 13:36:54
  参数: topic,主题
  参数: args,参数
  描述: Publish executes callback defined for a topic
*/
func (bus *EventBus) Publish(topic string, args ...interface{}) {
	var onceList []*eventHandler = nil
	//单次运行列表

	func() { //do publish
		bus.lock.RLock()
		defer bus.lock.RUnlock()

		handlers, ok := bus.handlers[topic]
		if !(ok && len(handlers) > 0) {
			return
		}

		for _, fn := range handlers {
			if fn.flagOnce {
				if onceList == nil {
					onceList = make([]*eventHandler, 0)
				}
				onceList = append(onceList, fn)
			}

			if fn.async {
				bus.rg.Run(func(params ...interface{}) {
					h, _ := params[0].(*eventHandler)
					p, _ := params[1].([]interface{})
					bus.doPublish(h, p...) //call
				}, fn, args)
			} else {
				bus.doPublish(fn, args...)
			}
		}
	}()

	if onceList != nil { //do once clear
		func() {
			bus.lock.Lock()
			defer bus.lock.Unlock()
			bus.deleteHandler(topic, onceList)
		}()
	}
}

/*doPublish 2022-08-19 13:11:11
  参数: handler,事件句柄
  参数: args,参数列表
  描述: 使用args调用handler
*/
func (bus *EventBus) doPublish(handler *eventHandler, args ...interface{}) {
	typ := handler.callBack.Type()
	params := make([]reflect.Value, len(args))
	//try fill args

	for i, v := range args {
		if v == nil {
			params[i] = reflect.New(typ.In(i)).Elem()
		} else {
			params[i] = reflect.ValueOf(v)
		}
	}

	handler.callBack.Call(params)
	//call
}

/*deleteHandler 2022-08-19 12:05:27
  参数: topic,主题名称
  参数: toDel,待删除句柄
  参数: fn,待删除方法
  描述: 删除topic主题指定的handler
*/
func (bus *EventBus) deleteHandler(topic string, toDel []*eventHandler, fn ...interface{}) error {
	if len(toDel) < 1 && fn == nil {
		return fmt.Errorf("znlib.EventBus: deleteHandler has invalid input.")
	}

	if fn != nil {
		for _, f := range fn {
			if f == nil { //nil参数
				continue
			}

			val := reflect.ValueOf(f)
			for _, h := range bus.handlers[topic] {
				if h.callBack.Type() == val.Type() && h.callBack.Pointer() == val.Pointer() {
					if toDel == nil {
						toDel = make([]*eventHandler, 0)
					}

					toDel = append(toDel, h)
					//fill list to delete
				}
			}
		}
	}

	if len(toDel) < 1 {
		return nil //empty list to delete
	}

	var (
		idx    int = 0
		exists bool
	)

	for _, h := range bus.handlers[topic] {
		exists = false
		for _, tmp := range toDel {
			if h == tmp {
				exists = true
				break
			}
		}

		if !exists { //无需删除,则移动索引
			bus.handlers[topic][idx] = h
			idx++
		}
	}

	bus.handlers[topic] = bus.handlers[topic][:idx]
	return nil
}

/*WaitAsync 2022-08-19 17:01:57
  描述: waits for all async callbacks to complete
*/
func (bus *EventBus) WaitAsync() {
	bus.rg.Wait()
}
