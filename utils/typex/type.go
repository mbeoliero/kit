package typex

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func ToAny[T any](value string) T {
	t, _ := ToAnyE[T](value)
	return t
}

// ToAnyE parses value into T. Integers are range checked for T's width; other types
// are decoded as JSON. An empty value yields T's zero value.
func ToAnyE[T any](value string) (T, error) {
	var t T
	if len(value) == 0 {
		return t, nil
	}
	var err error
	switch p := any(&t).(type) {
	case *string:
		*p = value
	case *int:
		*p, err = parseInt[int](value, strconv.IntSize)
	case *int8:
		*p, err = parseInt[int8](value, 8)
	case *int16:
		*p, err = parseInt[int16](value, 16)
	case *int32:
		*p, err = parseInt[int32](value, 32)
	case *int64:
		*p, err = parseInt[int64](value, 64)
	case *uint:
		*p, err = parseUint[uint](value, strconv.IntSize)
	case *uint8:
		*p, err = parseUint[uint8](value, 8)
	case *uint16:
		*p, err = parseUint[uint16](value, 16)
	case *uint32:
		*p, err = parseUint[uint32](value, 32)
	case *uint64:
		*p, err = parseUint[uint64](value, 64)
	case *float32:
		var v float64
		v, err = strconv.ParseFloat(value, 32)
		*p = float32(v)
	case *float64:
		*p, err = strconv.ParseFloat(value, 64)
	default:
		err = json.Unmarshal([]byte(value), &t)
	}
	return t, err
}

func parseInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value string, bits int) (T, error) {
	v, err := strconv.ParseInt(value, 10, bits)
	return T(v), err
}

func parseUint[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](value string, bits int) (T, error) {
	v, err := strconv.ParseUint(value, 10, bits)
	return T(v), err
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
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
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
		b, _ := json.Marshal(v)
		return string(b)
	}
}
