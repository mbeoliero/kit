package jsonx

import jsoniter "github.com/json-iterator/go"

func MarshalToString(v any) string {
	ret, _ := jsoniter.Marshal(v)
	return string(ret)
}
