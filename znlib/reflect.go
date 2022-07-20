/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-17 16:55:03
  描述: 反射相关操作
******************************************************************************/
package znlib

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/shopspring/decimal" //浮点数计算
	"reflect"
	"strconv"
	"sync"
)

/*IsIn 2022-07-17 17:25:38
  参数: val,值
  参数: array,数据
  描述: 判断val是否在array中,返回索引
*/
func IsIn(val interface{}, array interface{}) int {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(array)
		for idx := 0; idx < s.Len(); idx++ {
			if reflect.DeepEqual(val, s.Index(idx).Interface()) {
				return idx
			}
		}
	}

	return -1
}

/*IsNil 2022-07-17 16:56:52
  参数: val,值
  描述: 判断val是否为空
*/
func IsNil(val interface{}) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Pointer, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

/*Equal 2022-07-17 17:06:11
  参数: a,字节数组
  参数: b,字节数组
  描述: whether a and b are the same length and contain the same bytes
*/
func Equal(a, b interface{}) bool {
	if a == nil {
		return IsNil(b)
	}

	exp, ok := a.([]byte)
	if ok {
		act, ok := b.([]byte)
		if !ok {
			return false
		}
		return bytes.Equal(exp, act)
	}

	return reflect.DeepEqual(a, b)
}

/*ValueCompare 2022-07-17 19:05:58
  参数: a,数值
  参数: b,数值
  参数: relation,比较关系
  描述: 比较a、b的关系(大于 等于 小于...)
*/
func ValueCompare(a, b interface{}, relation ValueRelation) (ok bool) {
	var va, vb decimal.Decimal
	va, ok = ValueToDecimal(a)
	if !ok {
		return
	}

	vb, ok = ValueToDecimal(b)
	if !ok {
		return
	}

	switch relation {
	case ValEqual:
		return va.Equal(vb)
	case ValGreater:
		return va.GreaterThan(vb)
	case ValGreaterEqual:
		return va.GreaterThanOrEqual(vb)
	case ValLess:
		return va.LessThan(vb)
	case ValLessEqual:
		return va.LessThanOrEqual(vb)
	default:
		return false
	}
}

/*ValueToDecimal 2022-07-17 23:34:54
  参数: val,数值
  描述: 将val转为浮点数
*/
func ValueToDecimal(val interface{}) (decimal.Decimal, bool) {
	switch val.(type) {
	case float32:
		val32, ok := val.(float32)
		if ok {
			return decimal.NewFromFloat32(val32), true
		}
	case float64:
		val64, ok := val.(float64)
		if ok {
			return decimal.NewFromFloat(val64), true
		}
	case string:
		str, ok := val.(string)
		if ok {
			dec, err := decimal.NewFromString(str)
			if err == nil {
				return dec, true
			}
		}
	case int8, int16, int, int32, int64: //整数转换
		return decimal.NewFromInt(reflect.ValueOf(val).Int()), true
	case uint8, uint16, uint, uint32, uint64: //无符号整数
		return IsNumber(strconv.FormatUint(reflect.ValueOf(val).Uint(), 10))
	}

	return decimal.Zero, false
}

/*IsNumber 2022-07-17 23:15:52
  参数: str,字符串
  参数: isfloat,是否为浮点数
  描述: 判断str是否为数值
*/
func IsNumber(str string, isfloat ...bool) (decimal.Decimal, bool) {
	val, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero, false
	}

	if isfloat == nil {
		return val, true
	} else {
		return val, isfloat[0] == true || val.IsInteger()
	}
}

/*ReflectValue 2022-07-19 11:12:17
  参数: obj,对象
  描述: 返回obj的Value反射数据
*/
func ReflectValue(obj interface{}) reflect.Value {
	var val reflect.Value = reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		return val.Elem()
	} else {
		return val
	}
}

//StructFieldsWalker 结构体步进函数
type StructFieldsWalker = func(field reflect.StructField, value reflect.Value, level int) bool

/*WalkStruct 2022-07-19 11:23:14
  参数: obj,对象
  参数: sw,步进函数
  参数: level,当前层级(默认不传)
  描述: 检索obj的所有字段,并使用sw处理每个字段
*/
func WalkStruct(obj interface{}, sw StructFieldsWalker, level ...int) error {
	var curentLevel int = 1
	if level != nil {
		curentLevel = level[0]
		if curentLevel < 1 {
			curentLevel = 1
		}
	}

	if sw == nil { //no walker
		str := "znlib.WalkStruct: walker is nil"
		if curentLevel == 1 {
			Error(str)
		}
		return errors.New(str)
	}

	objValue := ReflectValue(obj)
	objType := objValue.Type()
	if objValue.Kind() != reflect.Struct { //invalid type
		str := fmt.Sprintf("znlib.WalkStruct: [%s] is not struct", objType.Name())
		if curentLevel == 1 {
			Error(str)
		}
		return errors.New(str)
	}

	var next bool
	fieldCount := objType.NumField()

	for nIdx := 0; nIdx < fieldCount; nIdx++ {
		field := objType.Field(nIdx)
		if field.IsExported() {
			fieldValue := objValue.Field(nIdx)
			next = sw(field, fieldValue, curentLevel)
			if next && fieldValue.Kind() == reflect.Struct {
				WalkStruct(fieldValue.Interface(), sw, curentLevel+1)
			}
		}
	}

	return nil
}

type structTags struct {
	typ      reflect.Type      //struct type
	key      string            //tag(key)名
	tagArray []string          //tag序列
	tags     map[string]string //k:字段名;v:tag值
}

var structTagsBuffer = struct {
	locker sync.RWMutex //同步锁定
	buffer []structTags //缓存
}{
	buffer: make([]structTags, 0),
}

/*getStructTags 2022-07-20 13:42:27
  参数: objType,对象类型
  参数: key,tag名
  参数: lock,是否锁定
  描述: 从缓存中检索类型为objType,tag名为key的索引
*/
func getStructTags(objType reflect.Type, key string, lock bool) (idx int) {
	if lock {
		structTagsBuffer.locker.RLock()
		//read lock
	}

	for idx = 0; idx < len(structTagsBuffer.buffer); idx++ {
		if structTagsBuffer.buffer[idx].typ == objType &&
			structTagsBuffer.buffer[idx].key == key {
			return idx
		}
	}

	if lock {
		structTagsBuffer.locker.RUnlock()
		//read unlock
	}

	return -1
}

/*setStructTags 2022-07-20 14:03:49
  参数: obj,对象
  参数: tags,缓存对象
  参数: deep,是否检索全部层级
  描述: 将obj中tag匹配的值存入tags缓存中
*/
func setStructTags(obj interface{}, tags *structTags, deep bool) error {
	var tag string
	err := WalkStruct(obj, func(field reflect.StructField, value reflect.Value, level int) bool {
		if field.Anonymous {
			return true
		}

		tag = field.Tag.Get(tags.key)
		if tag != "" {
			tags.tags[field.Name] = tag                //map
			tags.tagArray = append(tags.tagArray, tag) //array
		}

		return deep || level == 1 //检索1层或全部
	})

	if err == nil { //存入缓存
		structTagsBuffer.locker.Lock()
		//write lock
		if getStructTags(tags.typ, tags.key, false) < 0 {
			structTagsBuffer.buffer = append(structTagsBuffer.buffer, *tags)
		}

		structTagsBuffer.locker.Unlock()
		//write unlock
	}

	return err
}

/*StructTags 2022-07-19 12:34:35
  参数: obj,对象
  参数: key,Tag名
  参数: deep,是否检索全部层级
  描述: 获取obj包含key的字段名和tag值
*/
func StructTags(obj interface{}, key string, deep bool) (map[string]string, error) {
	objType := ReflectValue(obj).Type()
	idx := getStructTags(objType, key, true)
	if idx >= 0 {
		return structTagsBuffer.buffer[idx].tags, nil
	}

	var nTags = structTags{
		objType,
		key,
		make([]string, 0),
		make(map[string]string),
	}

	err := setStructTags(obj, &nTags, deep)
	//fill data
	return nTags.tags, err
}

/*StructTagList 2022-07-20 14:08:20
  参数: obj,obj,对象
  参数: key,key,Tag名
  参数: deep,deep,是否检索全部层级
  描述: 获取obj包含key的tag值列表,按obj字段的先后顺序排列
*/
func StructTagList(obj interface{}, key string, deep bool) ([]string, error) {
	objType := ReflectValue(obj).Type()
	idx := getStructTags(objType, key, true)
	if idx >= 0 {
		return structTagsBuffer.buffer[idx].tagArray, nil
	}

	var nTags = structTags{
		objType,
		key,
		make([]string, 0),
		make(map[string]string),
	}

	err := setStructTags(obj, &nTags, deep)
	//fill data
	return nTags.tagArray, err
}
