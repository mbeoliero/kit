package redisx

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// The TTL is applied only while the key has none, i.e. when INCRBY just created it, so later
// calls never extend it. A non-positive expire leaves the key without a TTL; EXPIRE 0 would
// delete it.
var incrByScript = redis.NewScript(`
local current = redis.call("INCRBY", KEYS[1], ARGV[1])
if tonumber(ARGV[2]) > 0 and redis.call("TTL", KEYS[1]) == -1 then
	redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return current
`)

func Incr(ctx context.Context, key string, expire time.Duration) (int64, error) {
	return IncrByClient(ctx, GlobalClient, key, expire)
}

func Decr(ctx context.Context, key string, expire time.Duration) (int64, error) {
	return DecrByClient(ctx, GlobalClient, key, expire)
}

func IncrByClient(ctx context.Context, cli redis.UniversalClient, key string, expire time.Duration) (int64, error) {
	return incrBy(ctx, cli, key, 1, expire)
}

func DecrByClient(ctx context.Context, cli redis.UniversalClient, key string, expire time.Duration) (int64, error) {
	return incrBy(ctx, cli, key, -1, expire)
}

func incrBy(ctx context.Context, cli redis.UniversalClient, key string, delta int64, expire time.Duration) (int64, error) {
	return incrByScript.Run(ctx, cli, []string{key}, delta, int64(expire/time.Second)).Int64()
}
