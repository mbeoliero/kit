package log

import (
	"context"
	"maps"
)

type fieldsKey string

// CustomFieldsKey stores the log fields carried by a context.
const CustomFieldsKey fieldsKey = "ctx_extra_data"

// TraceIDKey is the conventional field name for a trace ID.
const TraceIDKey = "trace_id"

// AppendLogExtras returns a context whose log fields include extra. The parent
// context's fields are never modified, so derived contexts may be used concurrently.
func AppendLogExtras(ctx context.Context, extra map[string]string) context.Context {
	if len(extra) == 0 {
		return ctx
	}
	fields := maps.Clone(GetAllCustomFields(ctx))
	if fields == nil {
		fields = make(map[string]string, len(extra))
	}
	maps.Copy(fields, extra)
	return context.WithValue(ctx, CustomFieldsKey, fields)
}

// AppendLogKv returns a context whose log fields include key=value. The parent
// context's fields are never modified.
func AppendLogKv(ctx context.Context, key, value string) context.Context {
	return AppendLogExtras(ctx, map[string]string{key: value})
}

// GetAllCustomFields returns the context's log fields. The map is shared with every
// context derived from ctx and must not be modified.
func GetAllCustomFields(ctx context.Context) map[string]string {
	fields, _ := ctx.Value(CustomFieldsKey).(map[string]string)
	return fields
}
