package redisx

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/mbeoliero/kit/utils/typex"
	"github.com/redis/go-redis/v9"
)

func Get[T any](ctx context.Context, key string) (T, error) {
	return GetByClient[T](ctx, GlobalClient, key)
}

func Set[T any](ctx context.Context, key string, value T, expire time.Duration) error {
	return SetByClient[T](ctx, GlobalClient, key, value, expire)
}

func Del(ctx context.Context, key string) error {
	return DelByClient(ctx, GlobalClient, key)
}

func GetByClient[T any](ctx context.Context, cli redis.UniversalClient, key string) (T, error) {
	var res T
	resStr, err := cli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return res, nil
		}
		return res, err
	}

	return typex.ToAnyE[T](resStr)
}

func SetByClient[T any](ctx context.Context, cli redis.UniversalClient, key string, value T, expire time.Duration) error {
	return cli.Set(ctx, key, typex.ToString(value), expire).Err()
}

func DelByClient(ctx context.Context, cli redis.UniversalClient, key string) error {
	return cli.Del(ctx, key).Err()
}
