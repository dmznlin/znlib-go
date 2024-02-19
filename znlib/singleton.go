// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-02-19 15:09:32
  描述: 实现一个应用只启动一个实例
******************************************************************************/
package znlib

import (
	"os"
)

// singleton 单实例
type singleton struct {
	isValid     bool     //互斥有效
	mutex       string   //互斥名称
	mutexHandle uintptr  //互斥句柄
	fileName    string   //互斥文件名
	fileHandle  *os.File //互斥文件句柄
}

// Singleton 实现单实例
var Singleton = &singleton{isValid: false}

// Lock 2024-02-19 16:17:33
/*
 参数: mutex,互斥量名称
 描述: 尝试锁定mutex,若成功则为首次启动
*/
func (st *singleton) Lock(mutex string) bool {
	if !st.isValid {
		st.mutex = mutex
		return mutexLock(st, "znlib.singleton.Lock")
	}

	return st.isValid
}

// Unlock 2024-02-19 16:18:43
/*
 描述: 释放互斥信号
*/
func (st *singleton) Unlock() {
	if st.isValid {
		st.isValid = false
		mutexUnlock(st, "znlib.singleton.Unlock")
	}

}
