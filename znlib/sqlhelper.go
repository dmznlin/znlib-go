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

/*SQLFields 2022-07-15 15:54:59
  参数: obj,struct结构体
  描述: 返回obj结构体中tag=db的字段名列表
*/
func SQLFields(obj interface{}) (fields []string) {
	tags, err := StructTags(obj, "db", true)
	if err != nil {
		return nil
	}

	var idx int = 0
	fields = make([]string, len(tags))
	for _, tag := range tags {
		fields[idx] = tag
		idx++
	}

	return
}

/*SQLFieldsJoin 2022-07-15 16:25:23
  参数: obj,struct结构体
  描述: 拼接字段名
*/
func SQLFieldsJoin(obj interface{}) string {
	fields := SQLFields(obj)
	if fields == nil {
		return "*"
	} else {
		return strings.Join(fields, ",")
	}
}

/*SQLInsert 2022-07-19 19:15:13
  参数: obj,struct结构体
  参数: fiedls...string,排除的字段 或 包含的字段
  描述: 使用obj构建insert sql语句
*/
func SQLInsert(obj interface{}, fiedls ...string) (sql string, err error) {
	defer DeferHandle(false, "SQLFields", func(err any) {
		if err != nil {
			err = errors.New(fmt.Sprintf("znlib.SQLInsert: %v", err))
		}
	})

	var (
		nBool   bool
		nTag    string
		nTable  string    = ""
		nDBType SqlDbType = SQLDB_mssql

		nPrefix  []string = make([]string, 0)
		nSuffix  []string = make([]string, 0)
		nInclude bool     = StrIn(SQLTag_Include, fiedls...)
	)

	for _, dt := range SQLDB_Types { //find database type
		if StrIn(dt, fiedls...) {
			nDBType = dt
			break
		}
	}

	err = WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if value.Kind() != reflect.Struct {
			if nTable == "" { //table
				nTag = field.Tag.Get(SQLTag_Table)
				if nTag != "" {
					nTable = nTag
				}
			}

			nTag = field.Tag.Get(SQLTag_DB)
			if nTag != "" { //field
				nBool = StrIn(nTag, fiedls...)
				if (nInclude && nBool) || (!nInclude && !nBool) {
					nPrefix = append(nPrefix, nTag)
					//field

					nSuffix = append(nSuffix, SQLValue(value.Interface(), nDBType))
					//value
				}
			}
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

	var strQuotes string = SqlQuotes_Single
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
