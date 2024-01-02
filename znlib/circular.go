// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-08-24 09:43:32
  描述: 支持FIFO、FILO的循环队列

  *.队列底层数据为双向链表,头尾闭合成环:
  	 <==> data <==> head <==> data <==> data <==> data
  	 ||											   ||
  	 |<== data <==>  data <==> tail <==> data <==> ||
  添加数据时: next顺时针
  读取数据时: FIFO为next顺时针,FILO为prior逆时针
******************************************************************************/
package znlib

import (
	"errors"
)

// CircularMode 队列模式
type CircularMode int8

const (
	Circular_FIFO         CircularMode = iota //队列模式: first in,first out
	Circular_FIFO_FixSize                     //固定大小队列,旧数据被覆盖
	Circular_FILO                             //栈模式: first in,last out
	Circular_FILO_FixSize                     //固定大小栈,旧数据被覆盖
)

// CircularMaxCap 队列最大容量
var CircularMaxCap = 1024

// circularItem 循环队列中的数据项
type circularData[T any] struct {
	data  T                //数据
	prior *circularData[T] //前项
	next  *circularData[T] //后项
}

// CircularQueue 循环队列
type CircularQueue[T any] struct {
	lock RWLocker         //同步锁
	mode CircularMode     //模式
	max  int              //最大容量
	cap  int              //当前容量
	num  int              //数据个数
	head *circularData[T] //队首数据
	tail *circularData[T] //队尾数据
}

// NewCircularQueue 2022-08-24 09:51:40
/*
 参数: mode,队列模式
 参数: cap,初始化大小
 参数: sync,需要同步锁
 参数: max,最大可容纳
 描述: 生成一个 T 类型的环形队列

 调用方法:
 queue := NewCircularQueue[int](Circular_FIFO, 0)
*/
func NewCircularQueue[T any](mode CircularMode, cap int, sync bool, max ...int) *CircularQueue[T] {
	if mode > Circular_FILO_FixSize {
		panic(errors.New("znlib.NewCircularQueue: invalid mode"))
	}

	if cap < 1 {
		cap = 3 //循环链中,最少3各元素
	}

	var maxNum int
	if max != nil && max[0] >= cap {
		maxNum = max[0]
	}

	if maxNum < cap {
		maxNum = CircularMaxCap
	}

	queue := CircularQueue[T]{
		tail: nil,
		mode: mode,
		cap:  cap,
		num:  0,
		max:  maxNum,
		lock: RWLocker{Enable: sync},
	}

	queue.head = &circularData[T]{prior: nil, next: nil}
	cd := queue.head
	//首元素

	for i := 1; i < cap; i++ {
		cd.next = &circularData[T]{prior: cd, next: nil} //链尾新建元素
		cd = cd.next
	}

	cd.next = queue.head
	queue.head.prior = cd //闭合成环
	return &queue
}

// Push 2022-08-24 09:57:39
/*
 参数: values,值列表
 描述: 添加一组值到队列中
*/
func (cq *CircularQueue[T]) Push(values ...T) error {
	if values == nil { //empty
		return errors.New("znlib.CircularQueue.Push: no value to push.")
	}

	cq.lock.Lock()
	defer cq.lock.Unlock()

	for _, val := range values {
		if cq.tail == nil { //首次添加
			cq.head.data = val
			cq.tail = cq.head
			cq.num++
			continue
		}

		if cq.tail.next == cq.head { //尾部没有空间
			if cq.mode == Circular_FIFO_FixSize || cq.mode == Circular_FILO_FixSize { //固定大小,覆盖旧数据
				cq.head = cq.head.next
				cq.tail = cq.tail.next
				cq.tail.data = val
			} else {
				if cq.num >= cq.max { //超出最大容量
					return errors.New("znlib.CircularQueue.Push: out of max capacity.")
				}

				cq.tail.next = &circularData[T]{prior: cq.tail, next: cq.tail.next} //插入新元素
				cq.tail = cq.tail.next
				cq.tail.data = val

				cq.cap++
				cq.num++
			}
		} else { //尾部有空间
			cq.tail = cq.tail.next
			cq.tail.data = val
			cq.num++
		}
	}

	return nil
}

// Pop 2022-08-24 12:06:12
/*
 参数: def,默认值
 描述: 取出队列中的值,若不存在则返回默认
*/
func (cq *CircularQueue[T]) Pop(def T) (value T, ok bool) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	ok = cq.tail != nil
	if !ok { //队列为空
		return def, ok
	}

	switch cq.mode {
	case Circular_FIFO, Circular_FIFO_FixSize: //先进先出
		value = cq.head.data
		if cq.head == cq.tail { //最后一个元素
			cq.tail = nil
		} else {
			cq.head = cq.head.next
		}
	case Circular_FILO, Circular_FILO_FixSize: //先进后出
		value = cq.tail.data
		if cq.head == cq.tail { //最后一个元素
			cq.tail = nil
		} else {
			cq.tail = cq.tail.prior
		}
	}

	cq.num--
	return value, ok
}

// MPop 2022-08-24 14:12:35
/*
 参数: num,个数
 描述: 取出多个元素
*/
func (cq *CircularQueue[T]) MPop(num int) (values []T) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	if cq.tail == nil || num <= 0 { //队列为空
		return []T{}
	}

	var idx int = 0
	values = make([]T, num)

	for num > 0 {
		switch cq.mode {
		case Circular_FIFO, Circular_FIFO_FixSize: //先进先出
			values[idx] = cq.head.data
			idx++

			if cq.head == cq.tail { //最后一个元素
				cq.tail = nil
			} else {
				cq.head = cq.head.next
			}
		case Circular_FILO, Circular_FILO_FixSize: //先进后出
			values[idx] = cq.tail.data
			idx++

			if cq.head == cq.tail { //最后一个元素
				cq.tail = nil
			} else {
				cq.tail = cq.tail.prior
			}
		}

		num--
		cq.num-- //元素计数
		if cq.tail == nil {
			break
		}
	}

	return values[:idx]
}

// Size 2022-08-24 14:09:27
/*
 返回: num,有效元素个数
 返回: cap,队列容量
 描述: 返回队列容量和元素个数
*/
func (cq *CircularQueue[T]) Size() (num, cap int) {
	return cq.num, cq.cap
}

// Walk 2022-08-24 15:44:14
/*
 参数: walk,遍历函数
 描述: 遍历队列中的元素

 walk func:
 参数: idx,元素索引
 参数: value,元素值
 参数: next,是否继续遍历

 例子:
 queue := NewCircularQueue[string](znlib.Circular_FILO, 0, 3)
 queue.Push("Hello", "World", "U", "Are", "Welcome") //value list

 queue.Walk(func(idx int, value string, next *bool) {
   t.Logf("walk val: %d,%s", idx, value)
   *next = value != "U"
 })
*/
func (cq *CircularQueue[T]) Walk(walk func(idx int, value T, next *bool)) {
	cq.lock.RLock()
	defer cq.lock.RUnlock()

	if cq.tail == nil { //队列为空
		return
	}

	var (
		idx  int  = 0
		next bool = true
	)

	switch cq.mode {
	case Circular_FIFO, Circular_FIFO_FixSize: //先进先出
		cd := cq.head
		for cd != nil {
			walk(idx, cd.data, &next) //callback
			if !next {
				break
			}

			if cd == cq.tail { //队尾
				cd = nil
			} else {
				cd = cd.next
				idx++
			}
		}
	case Circular_FILO, Circular_FILO_FixSize: //先进后出
		cd := cq.tail
		for cd != nil {
			walk(idx, cd.data, &next) //callback
			if !next {
				break
			}

			if cd == cq.head { //队首
				cd = nil
			} else {
				cd = cd.prior
				idx++
			}
		}
	}
}
