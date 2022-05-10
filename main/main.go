package main

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println(Copy("dmzn", 3, 3))
	Info("hello only")
	Info("hello with fields", logrus.Fields{"name": "dmzn", "age": 15}, logrus.Fields{"act": "eat"})
	Error("some error")
	Error("error with fields", logrus.Fields{"act": "run"})
}
