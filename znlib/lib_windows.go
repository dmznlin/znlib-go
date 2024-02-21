// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-02-21 15:09:17
  描述: 仅在win平台有效的代码
******************************************************************************/
package znlib

import (
	"syscall"
	"unsafe"
)

// mutexLock 2024-02-19 16:45:56
/*
 参数: st,单实例数据
 参数: caller,调用者
 描述: 创建一个互斥量,成功返回true
*/
func mutexLock(st *singleton, caller string) bool {
	var err error
	kernel := syscall.NewLazyDLL("kernel32.dll")
	st.mutexHandle, _, err = kernel.NewProc("CreateMutexA").Call(0, 1, uintptr(unsafe.Pointer(&st.mutex)))

	if err != syscall.Errno(0) {
		ErrorCaller(err, caller)
		return false
	}

	st.isValid = true
	return true
}

// mutexUnlock 2024-02-19 16:49:01
/*
 参数: st,单实例数据
 参数: caller,调用者
 描述: 释放互斥量
*/
func mutexUnlock(st *singleton, caller string) {
	err := syscall.CloseHandle(syscall.Handle(st.mutexHandle)) //关闭互斥句柄
	if err != nil {
		ErrorCaller(err, caller)
	}
}
