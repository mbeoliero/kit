package redisx

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func Incr(ctx context.Context, key string, expire time.Duration) (int64, error) {
	return IncrByClient(ctx, GlobalClient, key, expire)
}

func Decr(ctx context.Context, key string, expire time.Duration) (int64, error) {
	return DecrByClient(ctx, GlobalClient, key, expire)
}

func IncrByClient(ctx context.Context, cli redis.UniversalClient, key string, expire time.Duration) (int64, error) {
	script := redis.NewScript(`
    local current = redis.call("INCR", KEYS[1])
    if current == 1 then
        redis.call("EXPIRE", KEYS[1], ARGV[1])
    end
    return current
`)

	result, err := script.Run(ctx, cli, []string{key}, int(expire.Seconds())).Result()
	return result.(int64), err
}

func DecrByClient(ctx context.Context, cli redis.UniversalClient, key string, expire time.Duration) (int64, error) {
	script := redis.NewScript(`
	local current = redis.call("DECR", KEYS[1])
	if current == 0 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
`)

	result, err := script.Run(ctx, cli, []string{key}, int(expire.Seconds())).Result()
	return result.(int64), err
}
