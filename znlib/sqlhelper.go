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
    sql, err := SQLInsert(&user,
		func(field *StructFieldValue) (sqlVal string, done bool) {//构建回调函数
			if StrIn(field.StructField, "ID") {
				field.ExcludeMe = true //排除指定字段
				return "", true
			}

			if field.TableField == "age" {//设置特殊值
				return "age+1", true
			}

			return "", false
		})
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

/*SQLFields 2022-07-15 16:25:23
  参数: obj,struct结构体
  参数: exclude,排除字段
  描述: 拼接字段名
*/
func SQLFields(obj interface{}, exclude ...string) string {
	fields, err := StructTagList(obj, "db", true)
	if err == nil {
		for idx := 0; idx < len(fields); idx++ {
			if StrIn(fields[idx], exclude...) {
				fields = append(fields[:idx], fields[idx+1:]...)
				idx--
			}
		}

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
func SQLInsert(obj interface{}, getVal GetStructFieldValue, dbType ...SqlDbType) (sql string, err error) {
	defer DeferHandle(false, "znlib.SQLInsert", func(err any) {
		if err != nil {
			err = errors.New(fmt.Sprintf("znlib.SQLInsert: %v", err))
		}
	})

	var (
		nPrefix = make([]string, 0)
		nSuffix = make([]string, 0)

		done   bool
		sqlVal string
		nValue = StructFieldValue{
			DbType:    SQLDB_Default,
			TableName: "",
		}
	)

	if dbType != nil {
		nValue.DbType = dbType[0]
		//update db type
	}

	err = WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if nValue.TableName == "" {
			nValue.TableName = field.Tag.Get(SQLTag_Table)
			//get table name
		}

		nValue.TableField = field.Tag.Get(SQLTag_DB)
		if nValue.TableField != "" { //field
			if getVal != nil {
				nValue.StructField = field.Name
				nValue.StructValue = value.Interface()
				nValue.ExcludeMe = false

				sqlVal, done = getVal(&nValue)
				if done && nValue.ExcludeMe {
					return false
					//该字段已排除,不参与构建sql
				}
			} else {
				done = false
			}

			if !done { //默认取值
				sqlVal = SQLValue(value.Interface(), nValue.DbType)
				done = true
			}

			if done {
				nPrefix = append(nPrefix, nValue.TableField)
				//field
				nSuffix = append(nSuffix, sqlVal)
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

	if nValue.TableName == "" {
		str := fmt.Sprintf("znlib.SQLInsert: struct [%s] no [table] tag.", ReflectValue(obj).Type().Name())
		return "", errors.New(str)
	}

	sql = fmt.Sprintf("insert into %s(%s) values(%s)", nValue.TableName,
		strings.Join(nPrefix, ","), strings.Join(nSuffix, ","))
	return sql, nil
}

/*SQLUpdate 2022-07-20 17:24:20
  参数: obj,struct结构体
  参数: where,更新条件(可选空"")
  参数: getVal,获取字段值(可选nil)
  参数: DbType,数据库类型(默认不填写)
  描述: 使用obj构建update sql语句
*/
func SQLUpdate(obj interface{}, where string, getVal GetStructFieldValue, dbType ...SqlDbType) (sql string, err error) {
	defer DeferHandle(false, "znlib.SQLUpdate", func(err any) {
		if err != nil {
			err = errors.New(fmt.Sprintf("znlib.SQLUpdate: %v", err))
		}
	})

	var (
		done    bool
		sqlVal  string
		nFields = make([]string, 0)

		nValue = StructFieldValue{
			DbType:    SQLDB_Default,
			TableName: "",
		}
	)

	if dbType != nil {
		nValue.DbType = dbType[0]
		//update db type
	}

	err = WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if nValue.TableName == "" {
			nValue.TableName = field.Tag.Get(SQLTag_Table)
			//get table name
		}

		nValue.TableField = field.Tag.Get(SQLTag_DB)
		if nValue.TableField != "" { //field
			if getVal != nil {
				nValue.StructField = field.Name
				nValue.StructValue = value.Interface()
				nValue.ExcludeMe = false

				sqlVal, done = getVal(&nValue)
				if done && !nValue.ExcludeMe { //已处理且不排除
					nFields = append(nFields, nValue.TableField+"="+sqlVal)
				}
			} else {
				done = false
			}

			if !done { //默认取值
				nFields = append(nFields, nValue.TableField+"="+SQLValue(value.Interface(), nValue.DbType))
			}

			return false
			//带有db的字段,无需深层解析
		}
		return true
	})

	if err != nil {
		return "", err
	}

	if nValue.TableName == "" {
		str := fmt.Sprintf("znlib.SQLUpdate: struct [%s] no [table] tag.", ReflectValue(obj).Type().Name())
		return "", errors.New(str)
	}

	sql = fmt.Sprintf("update %s set %s%s", nValue.TableName,
		strings.Join(nFields, ","), StrIF(where == "", "", " where "+where))
	return sql, nil
}

type StructFieldValue struct {
	DbType     SqlDbType //数据库类型
	TableName  string    //数据库表名
	TableField string    //表字段名

	StructField string      //struct字段名
	StructValue interface{} //struct字段值
	ExcludeMe   bool        //排除该字段,不参与构建sql
}

/*GetStructFieldValue 获取field符合sql规范的值
  参数: field,struct字段数据
  返回: sqlVal,符合sql规范的值
  返回: done,是否已成功处理
*/
type GetStructFieldValue = func(field *StructFieldValue) (sqlVal string, done bool)

/*SQLValue 2022-07-19 21:09:13
  参数: value,数据
  参数: DbType,db类型
  描述: 转换value为字符串中的值
*/
func SQLValue(value interface{}, dbType SqlDbType) (val string) {
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
