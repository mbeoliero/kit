package log

import "context"

// AppendLogExtras 注意，此函数并非并发安全，请勿在初始化之外等常见进行写入
func AppendLogExtras(ctx context.Context, extra map[string]string) context.Context {
	extraData, ok := ctx.Value(CustomFieldsKey).(map[string]string)
	if !ok {
		extraData = make(map[string]string)
		ctx = context.WithValue(ctx, CustomFieldsKey, extraData)
	}
	for k, v := range extra {
		extraData[k] = v
	}
	return ctx
}

// AppendLogKv 注意，此函数并非并发安全，请勿在初始化之外等常见进行写入
func AppendLogKv(ctx context.Context, key, value string) context.Context {
	extraData, ok := ctx.Value(CustomFieldsKey).(map[string]string)
	if !ok {
		extraData = make(map[string]string)
		ctx = context.WithValue(ctx, CustomFieldsKey, extraData)
	}
	extraData[key] = value
	return ctx
}

func GetAllCustomFields(ctx context.Context) map[string]string {
	extraData, _ := ctx.Value(CustomFieldsKey).(map[string]string)
	return extraData
}
