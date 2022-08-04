package main

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type PS struct {
	F1 string    `db:"name"`
	F2 int       `db:"addr"`
	F3 time.Time `db:"phone"`
}

type userInfo struct {
	PS
	Id   int    `table:"sys_user" db:"r_id" `
	Age  int    `db:"u_age"`
	Name string `db:"u_name" table:"sys_user"`
}

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
	Info(StrDel("hello", 3, 6))

	var user = userInfo{
		PS:   PS{F1: "aa", F2: 10, F3: time.Now()},
		Id:   10,
		Name: "dmzn",
		Age:  5,
	}
	Info(SQLFields(&user))

	sql, err := SQLUpdate(&user, "id=2",
		func(field *StructFieldValue) (sqlVal string, done bool) { //构建回调函数
			if StrIn(field.StructField, "ID") { //排除指定字段
				field.ExcludeMe = true
				return "", true
			}

			if field.TableField == "u_age" { //设置特殊值
				return "u_age+1", true
			}

			return "", false
		})

	if err == nil {
		Info(sql)
	} else {
		Error(err.Error())
	}

	decimal.DivisionPrecision = 3
	v, ok := IsNumber("2.12345")
	if ok {
		Info(v.Div(decimal.NewFromInt32(3)).String())
	}

	/*
		WaitSystemExit(func() error {
			return errors.New("first cleaner")
		}, func() error {
			return errors.New("second cleaner")
		})
	*/
}
