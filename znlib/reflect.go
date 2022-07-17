/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-17 16:55:03
  描述: 反射相关操作
******************************************************************************/
package znlib

import (
	"bytes"
	"github.com/shopspring/decimal" //浮点数计算
	"reflect"
	"strconv"
)

/*Contains 2022-07-17 17:25:38
  参数: val,值
  参数: array,数据
  描述: 判断val是否在array中,返回索引
*/
func Contains(val interface{}, array interface{}) int {
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
