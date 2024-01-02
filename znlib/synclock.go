// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-10-29 16:38:47
  描述: 同步锁定
******************************************************************************/
package znlib

import "sync"

type RWLocker struct {
	locker sync.RWMutex //互斥量
	Enable bool         //是否有效
}

// RLock locks rw for reading.
func (lk *RWLocker) RLock() {
	if lk.Enable {
		lk.locker.RLock()
	}
}

// TryRLock tries to lock rw for reading and reports whether it succeeded.
func (lk *RWLocker) TryRLock() bool {
	if lk.Enable {
		return lk.locker.TryRLock()
	}

	return true
}

// RUnlock undoes a single RLock call;
func (lk *RWLocker) RUnlock() {
	if lk.Enable {
		lk.locker.RUnlock()
	}
}

// Lock locks rw for writing.
func (lk *RWLocker) Lock() {
	if lk.Enable {
		lk.locker.Lock()
	}
}

// TryLock tries to lock rw for writing and reports whether it succeeded.
func (lk *RWLocker) TryLock() bool {
	if lk.Enable {
		return lk.locker.TryLock()
	}

	return true
}

// Unlock unlocks rw for writing.
func (lk *RWLocker) Unlock() {
	if lk.Enable {
		lk.locker.Unlock()
	}
}
