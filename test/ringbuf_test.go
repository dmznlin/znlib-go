package test

import (
	"bytes"
	"encoding/binary"
	"github.com/dmznlin/znlib-go/znlib"
	"io"
	"strings"
	"testing"
)

func TestRingBuffer_interface(t *testing.T) {
	rb := znlib.NewRingBuffer(1, false)
	var _ io.Writer = rb
	var _ io.Reader = rb
	var _ io.StringWriter = rb
	var _ io.ByteReader = rb
	var _ io.ByteWriter = rb
}

func TestRingBuffer_Write(t *testing.T) {
	rb := znlib.NewRingBuffer(64, false)

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d.", rb.Length())
	}

	// check retrieve
	n, err := rb.Write([]byte(strings.Repeat("abcd", 2)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 8 {
		t.Fatalf("expect write 8 bytes but got %d", n)
	}
	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 2))) {
		t.Fatalf("expect 8 abcdabcd but got %s.", rb.Bytes())
	}
	rb.Retrieve(5)
	if rb.Length() != 3 {
		t.Fatalf("expect len 1 bytes but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("bcd", 1))) {
		t.Fatalf("expect 1 bcd but got %s. ", rb.Bytes())
	}
	_, err = rb.Write([]byte(strings.Repeat("abcd", 15)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if rb.Capacity() != 64 {
		t.Fatalf("expect capacity 64 bytes but got %d.", rb.Capacity())
	}
	if rb.Length() != 63 {
		t.Fatalf("expect len 63 bytes but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte("bcd"+strings.Repeat("abcd", 15))) {
		t.Fatalf("expect 63 ... but got %s. ", rb.Bytes())
	}
	rb.RetrieveAll()

	// write 4 * 4 = 16 bytes
	n, err = rb.Write([]byte(strings.Repeat("abcd", 4)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 16 {
		t.Fatalf("expect write 16 bytes but got %d", n)
	}
	if rb.Length() != 16 {
		t.Fatalf("expect len 16 bytes but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 4))) {
		t.Fatalf("expect 4 abcd but got %s.", rb.Bytes())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// write 48 bytes, should full
	n, err = rb.Write([]byte(strings.Repeat("abcd", 12)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 48 {
		t.Fatalf("expect write 48 bytes but got %d", n)
	}
	if rb.Length() != 64 {
		t.Fatalf("expect len 64 bytes but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 16))) {
		t.Fatalf("expect 16 abcd but got %s.", rb.Bytes())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// write more 4 bytes, should reject
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 1)))
	if rb.Length() != 68 {
		t.Fatalf("expect len 68 bytes but got %d.", rb.Length())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// reset this ringbuffer and set a long slice
	rb.Reset()
	n, _ = rb.Write([]byte(strings.Repeat("abcd", 20)))
	if n != 80 {
		t.Fatalf("expect write 80 bytes but got %d", n)
	}
	if rb.Length() != 80 {
		t.Fatalf("expect len 80 bytes but got %d.", rb.Length())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 20))) {
		t.Fatalf("expect 20 abcd but got %s.", rb.Bytes())
	}
}

func TestRingBuffer_Read(t *testing.T) {
	rb := znlib.NewRingBuffer(64, false)

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d.", rb.Length())
	}

	// read empty
	buf := make([]byte, 1024)
	n, err := rb.Read(buf)
	if err == nil {
		t.Fatalf("expect an error but got nil")
	}
	if err != znlib.ErrRingBufEmpty {
		t.Fatalf("expect ErrIsEmpty but got nil")
	}
	if n != 0 {
		t.Fatalf("expect read 0 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d.", rb.Length())
	}

	// write 16 bytes to read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 4)))
	n, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 16 {
		t.Fatalf("expect read 16 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d.", rb.Length())
	}

	// write long slice to  read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 20)))
	n, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 80 {
		t.Fatalf("expect read 80 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d.", rb.Length())
	}
}

func TestRingBuffer_Peek(t *testing.T) {
	rb := znlib.NewRingBuffer(16, false)

	buf := make([]byte, 8)
	// write 16 bytes to read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 4)))
	n, err := rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 8 {
		t.Fatalf("expect read 8 bytes but got %d", n)
	}
	if rb.Length() != 8 {
		t.Fatalf("expect len 8 bytes but got %d.", rb.Length())
	}

	first, end := rb.Peek(4)
	if len(first) != 4 {
		t.Fatalf("expect len 4 bytes but got %d", len(first))
	}
	if len(end) != 0 {
		t.Fatalf("expect len 0 bytes but got %d", len(end))
	}
	if !bytes.Equal(first, []byte(strings.Repeat("abcd", 1))) {
		t.Fatalf("expect abcd but got %s.", first)
	}

	_, _ = rb.Write([]byte("1234"))
	first, end = rb.Peek(10)
	if len(first) != 8 {
		t.Fatalf("expect len 8 bytes but got %d", len(first))
	}
	if len(end) != 2 {
		t.Fatalf("expect len 2 bytes but got %d", len(end))
	}
	if !bytes.Equal(first, []byte(strings.Repeat("abcd", 2))) {
		t.Fatalf("expect abcdabcd but got %s.", first)
	}
	if !bytes.Equal(end, []byte(strings.Repeat("12", 1))) {
		t.Fatalf("expect 12 but got %s.", end)
	}

	if !bytes.Equal(rb.Bytes(), []byte("abcdabcd1234")) {
		t.Fatalf("expect abcdabcd1234 but got %s.", rb.Bytes())
	}

	first, end = rb.PeekAll()
	if len(first) != 8 {
		t.Fatalf("expect len 8 bytes but got %d", len(first))
	}
	if len(end) != 4 {
		t.Fatalf("expect len 4 bytes but got %d", len(end))
	}
	if !bytes.Equal(first, []byte(strings.Repeat("abcd", 2))) {
		t.Fatalf("expect abcdabcd but got %s.", first)
	}
	if !bytes.Equal(end, []byte(strings.Repeat("1234", 1))) {
		t.Fatalf("expect 1234 but got %s.", end)
	}

	rb.Retrieve(10)
	if !bytes.Equal(rb.Bytes(), []byte("34")) {
		t.Fatalf("expect 34 but got %s. ", rb.Bytes())
	}
}

func TestRingBuffer_ByteInterface(t *testing.T) {
	rb := znlib.NewRingBuffer(2, false)

	// write one
	err := rb.WriteByte('a')
	if err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}
	if rb.Length() != 1 {
		t.Fatalf("expect len 1 byte but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte{'a'}) {
		t.Fatalf("expect a but got %s.", rb.Bytes())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// write to, isFull
	err = rb.WriteByte('b')
	if err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}
	if rb.Length() != 2 {
		t.Fatalf("expect len 2 bytes but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte{'a', 'b'}) {
		t.Fatalf("expect a but got %s.", rb.Bytes())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// write
	_ = rb.WriteByte('c')
	if rb.Length() != 3 {
		t.Fatalf("expect len 3 bytes but got %d.", rb.Length())
	}
	if rb.Capacity() != 4 {
		t.Fatalf("expect Capacity 3 bytes but got %d.", rb.Capacity())
	}

	if !bytes.Equal(rb.Bytes(), []byte{'a', 'b', 'c'}) {
		t.Fatalf("expect a but got %s.", rb.Bytes())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// read one
	b, err := rb.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}
	if b != 'a' {
		t.Fatalf("expect a but got %c.", b)
	}
	if rb.Length() != 2 {
		t.Fatalf("expect len 2 byte but got %d.", rb.Length())
	}

	if !bytes.Equal(rb.Bytes(), []byte{'b', 'c'}) {
		t.Fatalf("expect a but got %s.", rb.Bytes())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// read two, empty
	b, err = rb.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}
	if b != 'b' {
		t.Fatalf("expect b but got %c.", b)
	}
	if rb.Length() != 1 {
		t.Fatalf("expect len 1 byte but got %d.", rb.Length())
	}

	// read three, error
	_, _ = rb.ReadByte()
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 byte but got %d.", rb.Length())
	}

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// read four, error
	_, err = rb.ReadByte()
	if err == nil {
		t.Fatalf("expect ErrIsEmpty but got nil")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 byte but got %d.", rb.Length())
	}

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
}

func TestRingBuffer_VirtualXXX(t *testing.T) {
	rb := znlib.NewRingBuffer(10, false)

	_, err := rb.Write([]byte("abcd1234"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	buf := make([]byte, 4)
	_, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("abcd")) {
		t.Fatal()
	}

	buf = make([]byte, 2)
	_, err = rb.VirtualRead(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("12")) {
		t.Fatal()
	}
	if rb.Length() != 4 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}
	rb.VirtualFlush()
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}

	_, err = rb.VirtualRead(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("34")) {
		t.Fatal()
	}
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 0 {
		t.Fatal()
	}
	rb.VirtualRevert()
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}

}

func TestRingBuffer_PeekUintXX(t *testing.T) {
	rb := znlib.NewRingBuffer(1024, false)
	_ = rb.WriteByte(0x01)

	toWrite := make([]byte, 2)
	binary.BigEndian.PutUint16(toWrite, 100)
	_, _ = rb.Write(toWrite)

	toWrite = make([]byte, 4)
	binary.BigEndian.PutUint32(toWrite, 200)
	_, _ = rb.Write(toWrite)

	toWrite = make([]byte, 8)
	binary.BigEndian.PutUint64(toWrite, 300)
	_, _ = rb.Write(toWrite)

	if rb.Length() != 15 {
		t.Fatal()
	}

	v := rb.PeekUint8()
	if v != 0x01 {
		t.Fatal()
	}
	rb.Retrieve(1)

	v1 := rb.PeekUint16()
	if v1 != 100 {
		t.Fatal()
	}
	rb.Retrieve(2)

	v2 := rb.PeekUint32()
	if v2 != 200 {
		t.Fatal()
	}
	rb.Retrieve(4)

	v3 := rb.PeekUint64()
	if v3 != 300 {
		t.Fatal(v3)
	}
	rb.Retrieve(8)
}

func TestRingBuffer_Bytes(t *testing.T) {
	var (
		size = 64
		ring *znlib.RingBuffer
		n    int
		err  error
		data []byte
	)

	ring = znlib.NewRingBuffer(size, false)
	n, err = ring.Write([]byte(strings.Repeat("abcd", 2)))
	if n != 8 || err != nil {
		t.Fatal()
	}
	data = make([]byte, 4)
	n, err = ring.Read(data)
	if n != 4 || err != nil {
		t.Fatal()
	}
	n, err = ring.Write([]byte(strings.Repeat("efgh", 15)))
	if n != 60 || err != nil {
		t.Fatal()
	}

	except := strings.Repeat("abcd", 1) + strings.Repeat("efgh", 15)
	actual := string(ring.Bytes())
	if except != actual {
		t.Fatalf("except %s, but got %s", except, actual)
	}
}
