package jsonx

import "encoding/json"

// MarshalToString returns v as JSON, or "" when v cannot be encoded.
func MarshalToString(v any) string {
	ret, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(ret)
}
