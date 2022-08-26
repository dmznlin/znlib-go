package test

import (
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
	"unsafe"
)

func TestCircularInt(t *testing.T) {
	queue := znlib.NewCircularQueue[int](znlib.Circular_FILO, 1)
	queue.Push()                   //nil
	queue.Push(10, 20, 30, 40, 50) //value list

	num, cap := queue.Size()
	t.Logf("size int: %d num:%d cap:%d", unsafe.Sizeof(queue), num, cap)
	t.Logf("mult pop: %v", queue.MPop(3))

	val, ok := queue.Pop(-1)
	for ok {
		t.Logf("pop val: %d", val)
		val, ok = queue.Pop(-1)
	}

	queue.Push(1, 2, 3)
	val, _ = queue.Pop(-1)
	if val != 3 {
		t.Error("znlib.CircularQueue.pop wrong")
	}
}

func TestCircularString(t *testing.T) {
	queue := znlib.NewCircularQueue[string](znlib.Circular_FIFO_FixSize, 10)
	queue.Push("Hello", "World", "U", "Are", "Welcome") //value list

	num, cap := queue.Size()
	t.Logf("size string: %d num:%d cap:%d", unsafe.Sizeof(queue), num, cap)

	val, ok := queue.Pop("a")
	for ok {
		t.Logf("pop val: %s", val)
		val, ok = queue.Pop("a")
	}

	queue.Push("a", "b", "c")
	val, _ = queue.Pop("")
	if val != "a" {
		t.Error("znlib.CircularQueue.pop wrong")
	}
}

func TestCircularWalk(t *testing.T) {
	queue := znlib.NewCircularQueue[string](znlib.Circular_FILO, 0, 3)
	queue.Push("Hello", "World", "U", "Are", "Welcome") //value list

	queue.Walk(func(idx int, value string, next *bool) {
		t.Logf("walk val: %d,%s", idx, value)
		*next = value != "U"
	})
}
