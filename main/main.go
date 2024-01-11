package main

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	mt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

func main() {
	InitLib(nil, nil)
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
	Info(StrDel("hello", 3, 6))

	decimal.DivisionPrecision = 3
	v, ok := IsNumber("2.12345")
	if ok {
		Info(v.Div(decimal.NewFromInt32(3)).String())
	}

	WaitFor(100*time.Millisecond, func() bool {
		Info("waitfor")
		return false
	})

	Mqtt.Start(func(client mt.Client, message mt.Message) {
		Info(string(message.Topic()) + string(message.Payload()))
	})

	group := NewRoutineGroup()
	mt := func(args ...interface{}) {
	loop:
		for i := 0; i < 10; i++ {
			Mqtt.Publish("", []string{"aa", "bb", "cc"})

			select {
			case <-Application.Ctx.Done():
				Info("cancel")
				break loop
			default:
				time.Sleep(5 * time.Second)
			}
		}
	}

	group.Run(mt)
	group.Run(mt)

	Application.OnExit(func() {
		Info("i am exit")
		group.Wait()
	})

	WaitSystemExit(func() error {
		return errors.New("first cleaner")
	}, func() error {
		return errors.New("second cleaner")
	}, func() error {
		Mqtt.Stop()
		return nil
	})

}
