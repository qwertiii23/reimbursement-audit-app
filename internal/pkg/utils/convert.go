package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// ToString 将任意类型转换为字符串
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToInt 将任意类型转换为整数
func ToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, fmt.Errorf("nil value")
	}
	
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ToInt64 将任意类型转换为64位整数
func ToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("nil value")
	}
	
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// ToFloat64 将任意类型转换为浮点数
func ToFloat64(value interface{}) (float64, error) {
	if value == nil {
		return 0, fmt.Errorf("nil value")
	}
	
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ToBool 将任意类型转换为布尔值
func ToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ToJSON 将对象转换为JSON字符串
func ToJSON(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析对象
func FromJSON(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), obj)
}

// Copy 深拷贝对象
func Copy(src, dst interface{}) error {
	// TODO: 实现深拷贝逻辑
	return nil
}

// Clone 克隆对象
func Clone(src interface{}) interface{} {
	// TODO: 实现克隆逻辑
	return nil
}

// IsNil 检查对象是否为nil
func IsNil(value interface{}) bool {
	if value == nil {
		return true
	}
	
	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflectValue.IsNil()
	default:
		return false
	}
}

// IsZero 检查对象是否为零值
func IsZero(value interface{}) bool {
	if value == nil {
		return true
	}
	
	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.String:
		return reflectValue.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflectValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflectValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflectValue.Float() == 0
	case reflect.Bool:
		return !reflectValue.Bool()
	case reflect.Ptr, reflect.Interface:
		return reflectValue.IsNil()
	default:
		return reflectValue.IsZero()
	}
}

// Equal 比较两个对象是否相等
func Equal(a, b interface{}) bool {
	// TODO: 实现对象比较逻辑
	return false
}

// GetTypeName 获取类型名称
func GetTypeName(value interface{}) string {
	if value == nil {
		return "nil"
	}
	
	reflectType := reflect.TypeOf(value)
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	
	return reflectType.Name()
}

// GetFieldValue 获取结构体字段值
func GetFieldValue(obj interface{}, field string) interface{} {
	// TODO: 实现获取字段值逻辑
	return nil
}

// SetFieldValue 设置结构体字段值
func SetFieldValue(obj interface{}, field string, value interface{}) error {
	// TODO: 实现设置字段值逻辑
	return nil
}

// HasField 检查结构体是否有指定字段
func HasField(obj interface{}, field string) bool {
	// TODO: 实现检查字段逻辑
	return false
}