package main

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
)

func main() {
	fmt.Println(Copy("dmzn",3, 3))
	Logger.Info("hello, file")
	Logger.Warn("any warning")
}
