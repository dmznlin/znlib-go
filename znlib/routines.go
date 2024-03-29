// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-08-19 19:12:01
  描述: goroutine组
******************************************************************************/
package znlib

import (
	"github.com/dmznlin/znlib-go/znlib/copier"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// RoutineGroup routine组
type RoutineGroup struct {
	waitGroup sync.WaitGroup
}

// NewRoutineGroup 新建分组
func NewRoutineGroup() *RoutineGroup {
	return new(RoutineGroup)
}

// routineFunction routine回调函数
type routineFunction = func(args ...interface{})

// Run 2022-08-19 19:12:58
/*
 参数: fn,函数
 参数: arg,参数
 描述: 在routine中调用fn
*/
func (g *RoutineGroup) Run(fn routineFunction, args ...interface{}) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()
		fn(args...)
	}()
}

// RunSafe 2022-08-19 19:15:05
/*
 参数: fn,函数
 参数: arg,参数
 描述: 在routine中调用fn,捕捉异常.
*/
func (g *RoutineGroup) RunSafe(fn routineFunction, args ...interface{}) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()                           //2.done
		defer DeferHandle(false, "znlib.routines.RunSafe") //1.log first
		fn(args...)
	}()
}

// Wait 2022-08-19 19:15:38
/*
 描述: 等待routine执行完毕
*/
func (g *RoutineGroup) Wait() {
	g.waitGroup.Wait()
}

// WaitRun 2022-09-09 10:32:51
/*
 参数: timeout,超时间隔
 参数: fn,函数
 参数: args,参数
 描述: 执行fn参数,等待timeout后超时退出
*/
func (g *RoutineGroup) WaitRun(timeout time.Duration, fn routineFunction, args ...interface{}) error {
	done := make(chan error, 1)
	//用于接收异常
	go func() {
		defer DeferHandle(false, "znlib.routines.WaitRun", func(err error) {
			done <- err
		})

		fn(args...)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errors.New("znlib.routines.WaitRun: timeout.")
	}
}

//--------------------------------------------------------------------------------

// Waiter 等待对象
type Waiter[T any] struct {
	lock sync.Mutex //同步锁定
	val  *T         //等待返回值
	done chan bool  //等待结果,true为执行完毕
}

// NewWaiter 2024-02-23 19:26:23
/*
 参数: fn,异步函数
 描述: 创建一个等待对象
*/
func NewWaiter[T any](fn func() T) *Waiter[T] {
	waiter := &Waiter[T]{
		val:  nil,
		done: make(chan bool, 1),
		lock: sync.Mutex{},
	}

	if fn != nil { //执行异步操作
		go func() {
			defer DeferHandle(false, "znlib.routines.NewWaiter", func(err error) {
				if err != nil {
					waiter.Wakeup(nil, true)
				}
			})

			val := fn()
			waiter.Wakeup(&val, true)
		}()
	}

	return waiter
}

// WaitFor 2024-02-23 19:42:16
/*
 参数: timeout,超时时长;0无限等待
 描述: 等待并返回结果,true表示数据有效
*/
func (wt *Waiter[T]) WaitFor(timeout time.Duration) (result *T, ok bool) {
	if timeout < 1 { //等待直到唤醒
		ok = <-wt.done
		return wt.val, ok && wt.val != nil
	}

	ticker := time.After(timeout)
	for {
		select {
		case <-ticker:
			return nil, false
		case ok = <-wt.done:
			return wt.val, ok && wt.val != nil
		}
	}
}

// Wakeup 2024-02-23 19:49:11
/*
 参数: val,返回值
 参数: direct,直接传递结果
 描述: 唤醒等待,并返回result
*/
func (wt *Waiter[T]) Wakeup(result *T, direct ...bool) {
	wt.lock.Lock()
	defer wt.lock.Unlock()
	//多线程唤醒保护

	if len(wt.done) < cap(wt.done) { //缓冲有效(未满)
		cp := (direct != nil) && (!direct[0])
		//需要复制(保护)结果

		if result != nil && cp {
			var val T
			err := copier.Copy(&val, result)
			if err != nil {
				ErrorCaller(err, "znlib.routines.Wakeup")
				return
			}
			//复制内容(deep-copy),避免传递时外部修改
			result = &val
		}

		wt.val = result
		wt.done <- true
	}
}

// Reset 2024-02-27 16:11:11
/*
 描述: 重置信号状态,再次WaitFor前调用
*/
func (wt *Waiter[T]) Reset() {
	wt.lock.Lock()
	defer wt.lock.Unlock()
	//锁定: 只清理当前信号

	if len(wt.done) > 0 { //有唤醒信号
	loop:
		for {
			select {
			case <-wt.done:
				//移除唤醒信号
			default:
				break loop
			}
		}
	}
}
