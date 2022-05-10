package main

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
)

func main() {
	fmt.Println(Copy("dmzn", 3, 3))
	Info("hello only")
	Info("hello with fields", LogFields{"name": "dmzn", "age": 15}, LogFields{"act": "eat"})
	Error("some error")
	Error("error with fields", LogFields{"act": "run"})
}
