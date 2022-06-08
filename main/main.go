package main

import (
	"errors"
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	"strconv"
)

func main() {
	fmt.Println(StrCopy("dmzn", 3, 3))
	Info("hello only")
	Info("hello with fields", LogFields{"name": "dmzn", "age": 15}, LogFields{"act": "eat"})
	Error("some error")
	Error("error with fields", LogFields{"act": "run"})

	str := strconv.Itoa(StrPos("中文H测试", "h"))
	Info(str)
	Info(StrReplace("中文english混合测试", "虎", "li", "合", "试"))

	Info(DateTime2Str(Str2DateTime("2022-06-02")))
	WriteDefaultLog("hello")

	WaitSystemExit(func() error {
		return errors.New("first cleaner")
	}, func() error {
		return errors.New("second cleaner")
	})
}
