package jsonx

import (
	"github.com/bytedance/sonic"
)

func MarshalToString(v any) string {
	ret, _ := sonic.MarshalString(v)
	return ret
}
