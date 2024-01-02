// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-10-29 16:17:47
  描述: 自动扩容的循环缓冲区

  参考: https://github.com/Allenxuxu/ringbuffer
  作者: Allenxuxu
  协议: The MIT License (MIT)
******************************************************************************/
package znlib

import (
	"encoding/binary"
	"errors"
	"unsafe"
)

// ErrRingBufEmpty 缓冲区为空
var ErrRingBufEmpty = errors.New("znlib.RingBuffer: buffer is empty")

// RingBuffer 自动扩容循环缓冲区
type RingBuffer struct {
	RWLocker        //同步锁定
	buf      []byte //缓冲区
	initSize int    //初始大小
	size     int    //当前大小
	vr       int    //虚读索引: virtual read
	r        int    //读取索引: next position to read
	w        int    //写入索引: next position to write
	isEmpty  bool   //是否为空
}

// NewRingBuffer 2022-10-29 17:12:12
/*
 参数: size,
 参数: sync,
 描述: 初始RingBuffer缓冲区
*/
func NewRingBuffer(size int, sync bool) *RingBuffer {
	return &RingBuffer{
		RWLocker: RWLocker{Enable: sync},
		buf:      make([]byte, size),
		initSize: size,
		size:     size,
		isEmpty:  true,
	}
}

// WithData 2022-10-29 17:16:24
/*
 参数: data,切片数据
 描述: 将data转化为循环缓冲
*/
func (rb *RingBuffer) WithData(data []byte) {
	rb.r = 0
	rb.w = 0
	rb.vr = 0
	rb.isEmpty = false
	rb.size = len(data)
	rb.initSize = len(data)
	rb.buf = data
}

// VirtualFlush 2022-10-29 17:18:38
/*
 描述: 刷新虚读指针,使其与读指针同步
*/
func (rb *RingBuffer) VirtualFlush() {
	rb.r = rb.vr
	if rb.r == rb.w {
		rb.isEmpty = true
	}
}

// VirtualRevert 2022-10-29 17:19:32
/*
 描述: 还原虚读指针,使其与读指针同步
*/
func (rb *RingBuffer) VirtualRevert() {
	rb.vr = rb.r
}

// VirtualRead 2022-10-29 17:21:05
/*
 参数: data,数据缓存
 描述: 虚读数据到data中,长度为data空间大小
 备注: 不移动 read 指针，需要配合 VirtualFlush 和 VirtualRevert 使用
*/
func (rb *RingBuffer) VirtualRead(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	if rb.isEmpty {
		return 0, ErrRingBufEmpty
	}

	n = len(data)
	if rb.w > rb.vr {
		if n > rb.w-rb.vr {
			n = rb.w - rb.vr
		}

		copy(data, rb.buf[rb.vr:rb.vr+n])
		// move vr
		rb.vr = (rb.vr + n) % rb.size

		if rb.vr == rb.w {
			rb.isEmpty = true
		}
		return
	}

	if n > rb.size-rb.vr+rb.w {
		n = rb.size - rb.vr + rb.w
	}

	if rb.vr+n <= rb.size {
		copy(data, rb.buf[rb.vr:rb.vr+n])
	} else {
		// head
		copy(data, rb.buf[rb.vr:rb.size])
		// tail
		copy(data[rb.size-rb.vr:], rb.buf[0:n-rb.size+rb.vr])
	}

	// move vr
	rb.vr = (rb.vr + n) % rb.size
	return
}

// VirtualLength 2022-10-29 17:26:06
/*
 描述: 虚拟长度，虚读后剩余可读数据长度
*/
func (rb *RingBuffer) VirtualLength() int {
	if rb.w == rb.vr {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.vr {
		return rb.w - rb.vr
	}

	return rb.size - rb.vr + rb.w
}

// RetrieveAll 2022-10-29 17:32:54
/*
 描述: 回收所有缓存空间
*/
func (rb *RingBuffer) RetrieveAll() {
	rb.r = 0
	rb.w = 0
	rb.vr = 0
	rb.isEmpty = true
}

// Retrieve 2022-10-29 17:33:12
/*
 参数: len,回收大小
 描述: 回收长度为len的缓存空间
*/
func (rb *RingBuffer) Retrieve(len int) {
	if rb.isEmpty || len <= 0 {
		return
	}

	if len < rb.Length() {
		rb.r = (rb.r + len) % rb.size
		rb.vr = rb.r

		if rb.w == rb.r {
			rb.isEmpty = true
		}
	} else {
		rb.RetrieveAll()
	}
}

// Peek 2022-10-29 17:35:59
/*
 参数: len,长度
 描述: 读取len个字节的数据
*/
func (rb *RingBuffer) Peek(len int) (first []byte, end []byte) {
	if rb.isEmpty || len <= 0 {
		return
	}

	if rb.w > rb.r {
		if len > rb.w-rb.r {
			len = rb.w - rb.r
		}

		first = rb.buf[rb.r : rb.r+len]
		return
	}

	if len > rb.size-rb.r+rb.w {
		len = rb.size - rb.r + rb.w
	}

	if rb.r+len <= rb.size {
		first = rb.buf[rb.r : rb.r+len]
	} else {
		// head
		first = rb.buf[rb.r:rb.size]
		// tail
		end = rb.buf[0 : len-rb.size+rb.r]
	}
	return
}

// PeekAll 2022-10-31 16:53:00
/*
 描述: 读取所有数据
*/
func (rb *RingBuffer) PeekAll() (first []byte, end []byte) {
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		first = rb.buf[rb.r:rb.w]
		return
	}

	first = rb.buf[rb.r:rb.size]
	end = rb.buf[0:rb.w]
	return
}

// PeekUint8 2022-10-31 16:53:54
/*
 描述: 读取 uint8 类型的数据
*/
func (rb *RingBuffer) PeekUint8() uint8 {
	if rb.Length() < 1 {
		return 0
	}

	f, e := rb.Peek(1)
	if len(e) > 0 {
		return e[0]
	} else {
		return f[0]
	}
}

// PeekUint16 2022-10-31 16:54:19
/*
 描述: 读取 uint16 类型的数据
*/
func (rb *RingBuffer) PeekUint16() uint16 {
	if rb.Length() < 2 {
		return 0
	}

	f, e := rb.Peek(2)
	if len(e) > 0 {
		return binary.BigEndian.Uint16(rb.joinBytes(f, e))
	} else {
		return binary.BigEndian.Uint16(f)
	}
}

// PeekUint32 2022-10-31 16:56:57
/*
 描述: 读取 uint32 类型的数据
*/
func (rb *RingBuffer) PeekUint32() uint32 {
	if rb.Length() < 4 {
		return 0
	}

	f, e := rb.Peek(4)
	if len(e) > 0 {
		return binary.BigEndian.Uint32(rb.joinBytes(f, e))
	} else {
		return binary.BigEndian.Uint32(f)
	}
}

// PeekUint64 2022-10-31 16:58:00
/*
 描述: 读取 uint64 类型的数据
*/
func (rb *RingBuffer) PeekUint64() uint64 {
	if rb.Length() < 8 {
		return 0
	}

	f, e := rb.Peek(8)
	if len(e) > 0 {
		return binary.BigEndian.Uint64(rb.joinBytes(f, e))
	} else {
		return binary.BigEndian.Uint64(f)
	}
}

// Read 2022-10-31 16:58:00
/*
 参数: data,数据缓存
 描述: 读取数据到data中,长度为data空间大小
*/
func (rb *RingBuffer) Read(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	if rb.isEmpty {
		return 0, ErrRingBufEmpty
	}

	n = len(data)
	if rb.w > rb.r {
		if n > rb.w-rb.r {
			n = rb.w - rb.r
		}
		copy(data, rb.buf[rb.r:rb.r+n])
		// move readPtr
		rb.r = (rb.r + n) % rb.size
		if rb.r == rb.w {
			rb.isEmpty = true
		}
		rb.vr = rb.r
		return
	}

	if n > rb.size-rb.r+rb.w {
		n = rb.size - rb.r + rb.w
	}

	if rb.r+n <= rb.size {
		copy(data, rb.buf[rb.r:rb.r+n])
	} else {
		// head
		copy(data, rb.buf[rb.r:rb.size])
		// tail
		copy(data[rb.size-rb.r:], rb.buf[0:n-rb.size+rb.r])
	}

	// move readPtr
	rb.r = (rb.r + n) % rb.size
	if rb.r == rb.w {
		rb.isEmpty = true
	}
	rb.vr = rb.r
	return
}

// ReadByte 2022-10-31 17:04:46
/*
 描述: 读取单个字节
*/
func (rb *RingBuffer) ReadByte() (b byte, err error) {
	if rb.isEmpty {
		return 0, ErrRingBufEmpty
	}

	b = rb.buf[rb.r]
	rb.r++
	if rb.r == rb.size {
		rb.r = 0
	}

	if rb.w == rb.r {
		rb.isEmpty = true
	}

	rb.vr = rb.r
	return
}

// Write 2022-10-31 17:05:07
/*
 参数: data,数据缓存
 描述: 将data写入ringbuffer中
*/
func (rb *RingBuffer) Write(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	n = len(data)
	free := rb.free()
	if free < n {
		rb.enlargeSpace(n - free)
	}

	if rb.w >= rb.r {
		if rb.size-rb.w >= n {
			copy(rb.buf[rb.w:], data)
			rb.w += n
		} else {
			copy(rb.buf[rb.w:], data[:rb.size-rb.w])
			copy(rb.buf[0:], data[rb.size-rb.w:])
			rb.w += n - rb.size
		}
	} else {
		copy(rb.buf[rb.w:], data)
		rb.w += n
	}

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false
	return
}

// WriteByte 2022-10-31 17:18:51
/*
 参数: b,字节
 描述: 写入单字节
*/
func (rb *RingBuffer) WriteByte(b byte) error {
	if rb.free() < 1 {
		rb.enlargeSpace(1)
	}

	rb.buf[rb.w] = b
	rb.w++

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false
	return nil
}

// Length 2022-10-31 17:19:39
/*
 描述: ringbuffer有效数据长度
*/
func (rb *RingBuffer) Length() int {
	if rb.w == rb.r {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.r {
		return rb.w - rb.r
	}

	return rb.size - rb.r + rb.w
}

// Capacity 2022-10-31 17:20:29
/*
 描述: ringbuffer容量大小
*/
func (rb *RingBuffer) Capacity() int {
	return rb.size
}

// WriteString 2022-10-31 17:21:07
/*
 参数: s,字符串
 描述: 将字符串写入ringbuffer
*/
func (rb *RingBuffer) WriteString(s string) (n int, err error) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return rb.Write(*(*[]byte)(unsafe.Pointer(&h)))
}

// Bytes 2022-10-29 17:41:32
/*
 描述: 返回所有可读数据，此操作不会移动读指针，仅仅是拷贝全部数据
*/
func (rb *RingBuffer) Bytes() (buf []byte) {
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		buf = make([]byte, rb.w-rb.r)
		copy(buf, rb.buf[rb.r:rb.w])
		return
	}

	buf = make([]byte, rb.size-rb.r+rb.w)
	copy(buf, rb.buf[rb.r:rb.size])
	copy(buf[rb.size-rb.r:], rb.buf[0:rb.w])
	return
}

// IsFull 2022-10-29 17:40:20
/*
 描述: 缓冲区是否已满
*/
func (rb *RingBuffer) IsFull() bool {
	return !rb.isEmpty && rb.w == rb.r
}

// IsEmpty 2022-10-29 17:39:35
/*
 描述: 缓冲区是否为空
*/
func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

// Reset 2022-10-29 17:36:52
/*
 描述: 将缓存缩容回初始化时的大小
*/
func (rb *RingBuffer) Reset() {
	rb.r = 0
	rb.vr = 0
	rb.w = 0
	rb.isEmpty = true

	if rb.size > rb.initSize {
		rb.buf = make([]byte, rb.initSize)
		rb.size = rb.initSize
	}
}

// grow 2022-10-29 17:53:21
/*
 参数: cap,容量大小
 描述: 计算容量为cap时,需要扩容的值(参考切片append策略)
*/
func (rb *RingBuffer) grow(cap int) int {
	newcap := rb.size
	doublecap := newcap + newcap

	if cap > doublecap {
		newcap = cap
	} else {
		if rb.size < 1024 {
			newcap = doublecap
		} else {
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			if newcap <= 0 {
				newcap = cap
			}
		}
	}
	return newcap
}

// enlargeSpace 2022-10-29 17:48:16
/*
 参数: len,需扩容大小
 描述: 对ringbuffer扩容len(至少)
*/
func (rb *RingBuffer) enlargeSpace(len int) {
	vlen := rb.VirtualLength()
	newSize := rb.grow(rb.size + len)
	newBuf := make([]byte, newSize)
	oldLen := rb.Length()
	_, _ = rb.Read(newBuf)

	rb.w = oldLen
	rb.r = 0
	rb.vr = oldLen - vlen
	rb.size = newSize
	rb.buf = newBuf
}

// free 2022-10-29 17:48:16
/*
 描述: 缓冲区的可用空间大小
*/
func (rb *RingBuffer) free() int {
	if rb.w == rb.r {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}

	return rb.size - rb.w + rb.r
}

// free 2022-10-29 18:05:41
/*
 参数: first,从读索引开始的数据
 参数: end,从0索引开始的数据
 描述: 将缓冲区尾部的数据 和 头部数据合并在一起
*/
func (rb *RingBuffer) joinBytes(first, end []byte) []byte {
	buf := make([]byte, len(first)+len(end))
	_ = copy(buf, first)
	_ = copy(buf[len(first):], end)
	return buf
}
