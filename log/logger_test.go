package log

import (
	"context"
	"testing"
)

func TestInfo(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "11223344556677889900")
	CtxInfo(ctx, "no extra")
	ctx = AppendLogKv(ctx, "user_id", "1234")
	CtxInfo(ctx, "test %s", "abc")
	Info("test %s", "abc")
	ctx = AppendLogExtras(ctx, map[string]string{"user_id": "aaa", "app_id": "2"})
	CtxInfo(ctx, "test app_id %s", "abc")
	Info("test app_id %s", "abc")

}
