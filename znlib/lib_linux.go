// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-02-19 16:42:12
  描述: 仅在linux平台有效的代码
******************************************************************************/
package znlib

import (
	"fmt"
	"os"
)

// mutexLock 2024-02-19 16:45:56
/*
 参数: st,单实例数据
 参数: caller,调用者
 描述: 创建一个互斥量,成功返回true
*/
func mutexLock(st *singleton, caller string) bool {
	//进程PID文件
	st.fileName = FixPathVar(fmt.Sprintf("$path/%s.pid", st.mutex))

	_, err := os.Stat(st.fileName)
	if err == nil { //进程已经存在
		Error(caller + ": delete file <" + st.fileName + "> and run again")
		return false
	}

	//创建进程PID文件
	st.fileHandle, err = os.OpenFile(st.fileName, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
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
	err := st.fileHandle.Close()
	if err != nil {
		ErrorCaller(err, caller)
	}
	// 删除该文件
	err = os.Remove(st.fileName)
	if err != nil {
		ErrorCaller(err, caller)
	}
}
