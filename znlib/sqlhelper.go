/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-15 15:53:40
  描述: 数据库sql相关函数

备注:
  1.构建SQL的struct结构示例:
	var user struct {
		ID   int    `db:"id"` table:"sys_user"
		Name string `db:"name"`
		Age  int    `db:"age"`
	}
  2.构建SQL语句示例:
    sql, err := SQLInsert(&user, SQLTag_Include, SQLDB_mssql, "ID", "name")
    *.SQLTag_Include,SQLTag_Exclude: 构建时只包含/需排除 指定的字段
    *.字段可以是 struct结构体的字段,或 数据库字段名
    *.SQLDB_mssql: 构建符合mssql规范的字符串
******************************************************************************/
package znlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/*SQLFieldsJoin 2022-07-15 16:25:23
  参数: obj,struct结构体
  描述: 拼接字段名
*/
func SQLFieldsJoin(obj interface{}) string {
	fields, err := StructTagList(obj, "db", true)
	if err == nil {
		return strings.Join(fields, ",")
	} else {
		return "*"
	}
}

/*SQLInsert 2022-07-19 19:15:13
  参数: obj,struct结构体
  参数: fields,排除的字段 或 包含的字段
  描述: 使用obj构建insert sql语句
*/
func SQLInsert(obj interface{}, fields ...string) (sql string, err error) {
	defer DeferHandle(false, "znlib.SQLInsert", func(err any) {
		if err != nil {
			err = errors.New(fmt.Sprintf("znlib.SQLInsert: %v", err))
		}
	})

	var (
		nBool   bool
		nTag    string
		nTable  = ""
		nDBType = SQLDB_Default

		nPrefix  = make([]string, 0)
		nSuffix  = make([]string, 0)
		nInclude = StrIn(SQLTag_Include, fields...)
	)

	for _, dt := range SQLDB_Types { //find database type
		if StrIn(dt, fields...) {
			nDBType = dt
			break
		}
	}

	err = WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if nTable == "" { //table
			nTag = field.Tag.Get(SQLTag_Table)
			if nTag != "" {
				nTable = nTag
			}
		}

		nTag = field.Tag.Get(SQLTag_DB)
		if nTag != "" { //field
			nBool = StrIn(field.Name, fields...) || StrIn(nTag, fields...)
			//struct field or db field

			if (nInclude && nBool) || (!nInclude && !nBool) {
				nPrefix = append(nPrefix, nTag)
				//field
				nSuffix = append(nSuffix, SQLValue(value.Interface(), nDBType))
				//value
			}

			return false
			//带有db的字段,无需深层解析
		}
		return true
	})

	if err != nil {
		return "", err
	}

	if nTable == "" {
		str := fmt.Sprintf("znlib.SQLInsert: struct [%s] no [table] tag.", ReflectValue(obj).Type().Name())
		return "", errors.New(str)
	}

	sql = fmt.Sprintf("insert into %s(%s) values(%s)", nTable,
		strings.Join(nPrefix, ","), strings.Join(nSuffix, ","))
	return sql, nil
}

/*SQLUpdate 2022-07-20 17:24:20
  参数: obj,struct结构体
  参数: where,更新条件
  参数: fields,排除的字段 或 包含的字段
  描述: 使用obj构建update sql语句
*/
func SQLUpdate(obj interface{}, where string, fields ...string) (sql string, err error) {
	defer DeferHandle(false, "znlib.SQLUpdate", func(err any) {
		if err != nil {
			err = errors.New(fmt.Sprintf("znlib.SQLUpdate: %v", err))
		}
	})

	var (
		nBool   bool
		nTag    string
		nTable  = ""
		nDBType = SQLDB_Default

		nFields  = make([]string, 0)
		nInclude = StrIn(SQLTag_Include, fields...)
	)

	for _, dt := range SQLDB_Types { //find database type
		if StrIn(dt, fields...) {
			nDBType = dt
			break
		}
	}

	err = WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if nTable == "" { //table
			nTag = field.Tag.Get(SQLTag_Table)
			if nTag != "" {
				nTable = nTag
			}
		}

		nTag = field.Tag.Get(SQLTag_DB)
		if nTag != "" { //field
			nBool = StrIn(field.Name, fields...) || StrIn(nTag, fields...)
			//struct field or db field

			if (nInclude && nBool) || (!nInclude && !nBool) {
				nFields = append(nFields, nTag+"="+SQLValue(value.Interface(), nDBType))
				//field = value
			}

			return false
			//带有db的字段,无需深层解析
		}
		return true
	})

	if err != nil {
		return "", err
	}

	if nTable == "" {
		str := fmt.Sprintf("znlib.SQLUpdate: struct [%s] no [table] tag.", ReflectValue(obj).Type().Name())
		return "", errors.New(str)
	}

	sql = fmt.Sprintf("update %s set %s%s", nTable,
		strings.Join(nFields, ","), StrIF(where == "", "", " where "+where))
	return sql, nil
}

/*SQLValue 2022-07-19 21:09:13
  参数: value,数据
  参数: db,db类型
  描述: 转换value为字符串中的值
*/
func SQLValue(value interface{}, db SqlDbType) (val string) {
	val = ""
	if value == nil {
		return val
	}

	var strQuotes = SqlQuotes_Single
	//字符串引号

	switch value.(type) {
	case string:
		val = strQuotes + value.(string) + strQuotes
	case []byte:
		val = strQuotes + string(value.([]byte)) + strQuotes
	case int8, int16, int32, int, int64:
		val = strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
	case uint8, uint16, uint32, uint, uint64:
		val = strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
	case time.Time:
		tm, _ := value.(time.Time)
		val = strQuotes + DateTime2Str(tm) + strQuotes
	case float32:
		ft := value.(float32)
		val = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case float64:
		ft := value.(float64)
		val = strconv.FormatFloat(ft, 'f', -1, 64)
	default:
		newValue, _ := json.Marshal(value)
		val = strQuotes + string(newValue) + strQuotes
	}
	return val
}
