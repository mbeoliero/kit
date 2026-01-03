package typex

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/bytedance/sonic"
)

func ToAny[T any](value string) T {
	t, _ := ToAnyE[T](value)
	return t
}

func ToAnyE[T any](value string) (T, error) {
	var t T
	var err error
	if len(value) == 0 {
		return t, err
	}

	switch any(t).(type) {
	case string:
		t = any(value).(T)
	case int:
		var v int
		v, err = strconv.Atoi(value)
		t = any(v).(T)
	case int8:
		var v int
		v, err = strconv.Atoi(value)
		t = any(int8(v)).(T)
	case int16:
		var v int
		v, err = strconv.Atoi(value)
		t = any(int16(v)).(T)
	case int32:
		var v int
		v, err = strconv.Atoi(value)
		t = any(int32(v)).(T)
	case int64:
		var v int64
		v, err = strconv.ParseInt(value, 10, 64)
		t = any(v).(T)
	case uint:
		var v int
		v, err = strconv.Atoi(value)
		t = any(uint(v)).(T)
	case uint8:
		var v int
		v, err = strconv.Atoi(value)
		t = any(uint8(v)).(T)
	case uint16:
		var v int
		v, err = strconv.Atoi(value)
		t = any(uint16(v)).(T)
	case uint32:
		var v int
		v, err = strconv.Atoi(value)
		t = any(uint32(v)).(T)
	case uint64:
		var v int
		v, err = strconv.Atoi(value)
		t = any(uint64(v)).(T)
	case float32:
		var v float64
		v, err = strconv.ParseFloat(value, 64)
		t = any(float32(v)).(T)
	case float64:
		var v float64
		v, err = strconv.ParseFloat(value, 64)
		t = any(v).(T)
	case []any:
		var v []any
		err = sonic.UnmarshalString(value, &v)
		t = any(v).(T)
	case map[string]any:
		var v map[string]any
		err = sonic.UnmarshalString(value, &v)
		t = any(v).(T)
	default:
		err = sonic.UnmarshalString(value, &t)
	}
	return t, err
}

func ToString(value any) string {
	switch v := value.(type) {
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
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
	case uintptr:
		return strconv.FormatUint(uint64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case []byte:
		return string(v)
	case error:
		return v.Error()
	default:
		rt := reflect.TypeOf(value)
		switch rt.Kind() {
		case reflect.Bool:
			return strconv.FormatBool(reflect.ValueOf(value).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(reflect.ValueOf(value).Float(), 'f', -1, 64)
		case reflect.String:
			return reflect.ValueOf(value).String()
		default:
		}
		s, _ := sonic.MarshalString(v)
		return s
	}
}
