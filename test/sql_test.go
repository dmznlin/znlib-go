package test

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	_ "github.com/mattn/go-adodb"
	"testing"
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

var user = userInfo{
	PS:   PS{F1: "aa", F2: 10, F3: time.Now()},
	Id:   10,
	Name: "dmzn",
	Age:  5,
}

func TestSQLJoin(t *testing.T) {
	str := SQLFields(&user, "u_name")
	if str != "name,addr,phone,r_id,u_age" {
		t.Errorf("znlib.SQLFieldsJoin error")
	}
}

func TestGetDB(t *testing.T) {
	rg := NewRoutineGroup()
	rg.Run(func(arg ...interface{}) {
		for i := 0; i < 3; i++ {
			conn, err := DBManager.GetDB("mssql_main")
			if err == nil {
				var users = struct {
					Name string `db:"U_Name"`
					ID   string `db:"U_Account"`
					Mail string `db:"U_Mail"`
				}{}

				err = conn.Get(&users, fmt.Sprintf("select %s from Sys_Users", SQLFields(&users)))
				if err == nil {
					Info(fmt.Sprintf("User:%s ID:%s Mail:%s", users.Name, users.ID, users.Mail))
				} else {
					Error(err)
				}
			} else {
				Error(err)
			}

			time.Sleep(1 * time.Second)
		}
	})
	/*
		rg.Run(func() {
			time.Sleep(1 * time.Second)
			DBManager.UpdateDSN("mssql_main", "hello")
		})
	*/
	rg.Wait()
}

func TestSQLUpdate(t *testing.T) {
	sql, err := SQLUpdate(&user, "id=2",
		func(field *StructFieldValue) (sqlVal string, done bool) { //构建回调函数
			if StrIn(field.StructField, "ID") {                    //排除指定字段
				field.ExcludeMe = true
				return "", true
			}

			if field.TableField == "u_age" { //设置特殊值
				return "u_age+1", true
			}

			return "", false
		})

	if err == nil {
		t.Log(sql)
	} else {
		t.Error(err)
	}
}

type typeA struct{}

type typeB struct {
	typeA //匿名嵌套
}

func (a typeA) Msg() {
	Info("ta.msg")
}

func TestTrans(t *testing.T) {
	rg := NewRoutineGroup()
	rg.Run(func(arg ...interface{}) {
		conn, err := DBManager.GetDB("mssql_main")
		if err != nil {
			t.Error(err)
		}

		_, err = conn.Begin()
		if err != nil {
			t.Error(err)
		}
	})

	rg.Wait()
}
