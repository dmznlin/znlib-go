package test

import (
	"encoding/binary"
	"encoding/json"
	"github.com/dmznlin/znlib-go/znlib/restruct"
	"testing"
)

// TestStruct 2024-11-15 16:11:22
/*
 参数: t,
 描述: 使用restruct将 strut 转为 []byte json
*/
func TestStruct(t *testing.T) {
	var first, next struct {
		ID   int32  `struct:"int32:16" `
		Name string `struct:"[10]byte" `
		Addr string `struct:"[20]byte"`
	}

	first.ID = 1001
	first.Name = "abcd"
	first.Addr = "1234"

	buf, err := restruct.Pack(binary.BigEndian, &first)
	if err != nil {
		t.Fatal(err)
	}

	size, err := restruct.SizeOf(&first)
	if err != nil {
		t.Fatal(err)
	}

	if len(buf) != size {
		t.Fatal("length mismatch")
	}

	err = restruct.Unpack(buf, binary.BigEndian, &next)
	if err != nil {
		t.Fatal(err)
	}

	str, err := json.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("buf: %v \njson: %s", buf, str)
	if next.Name != first.Name {
		t.Fatal("name mismatch")
	}
}
