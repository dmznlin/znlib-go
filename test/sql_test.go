package test

import (
	"fmt"
	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/dmznlin/znlib-go/znlib/threading"
	_ "github.com/mattn/go-adodb"
	"testing"
	"time"
)

func TestSQLJoin(t *testing.T) {
	var user struct {
		Id   int    `db:"r_id"`
		Name string `db:"u_name"`
		Age  int    `db:"u_age"`
	}

	str := SQLFields(&user, "u_name")
	if str != "r_id,u_age" {
		t.Errorf("znlib.SQLFieldsJoin error")
	}
}

func TestGetDB(t *testing.T) {
	DBManager.LoadConfig("D:\\Program Files\\MyVCL\\go\\znlib-go\\main\\bin\\db.ini")
	rg := threading.NewRoutineGroup()
	rg.Run(func() {
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

	rg.Run(func() {
		time.Sleep(1 * time.Second)
		DBManager.UpdateDSN("mssql_main", "hello")
	})

	rg.Wait()
}
