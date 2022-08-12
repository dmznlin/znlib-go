package test

import (
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestSnowflake(t *testing.T) {
	for i := 0; i < 10; i++ {
		id, err := znlib.SnowflakeID.NextID()
		if err == nil {
			znlib.Info(id)
		} else {
			t.Error(err)
		}

		str, err := znlib.SnowflakeID.NextStr()
		if err == nil {
			znlib.Info(str)
		} else {
			t.Error(err)
		}

		str, err = znlib.SnowflakeID.NextStr(true)
		if err == nil {
			znlib.Info(str)
		} else {
			t.Error(err)
		}

		id = znlib.SerialID.NextID()
		znlib.Info(id)

		str, err = znlib.SerialID.NextStr(true)
		if err == nil {
			znlib.Info(str)
		} else {
			t.Error(err)
		}

		str = znlib.SerialID.TimeID()
		znlib.Info(str)
	}
}
