// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-08-19 19:12:01
  描述: goroutine组
******************************************************************************/
package znlib

import (
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
