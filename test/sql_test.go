package test

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestSQLJoin(t *testing.T) {
	var user struct {
		Id   int    `db:"r_id"`
		Name string `db:"u_name"`
		Age  int    `db:"u_age"`
	}

	str := SQLFieldsJoin(&user)
	if str != "r_id,u_name,u_age" {
		t.Errorf("znlib.SQLFieldsJoin error")
	}
}
