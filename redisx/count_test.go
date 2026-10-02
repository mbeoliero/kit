package redisx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestIncrReturnsErrorWhenRedisUnavailable(t *testing.T) {
	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})
	defer cli.Close()
	if _, err := IncrByClient(t.Context(), cli, "k", 0); err == nil {
		t.Fatal("expected a connection error")
	}
	if _, err := DecrByClient(t.Context(), cli, "k", 0); err == nil {
		t.Fatal("expected a connection error")
	}
}

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("KIT_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("KIT_TEST_REDIS_ADDR is not set")
	}
	cli := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = cli.Close() })
	return cli
}

func TestIncrAppliesTtlOnlyOnCreate(t *testing.T) {
	cli := testRedis(t)
	ctx := t.Context()
	key := "kit:test:incr:" + t.Name()
	t.Cleanup(func() { cli.Del(context.Background(), key) })

	if v, err := IncrByClient(ctx, cli, key, time.Minute); err != nil || v != 1 {
		t.Fatalf("first incr = %d, %v", v, err)
	}
	cli.Expire(ctx, key, 10*time.Second)
	if v, err := IncrByClient(ctx, cli, key, time.Minute); err != nil || v != 2 {
		t.Fatalf("second incr = %d, %v", v, err)
	}
	if ttl := cli.TTL(ctx, key).Val(); ttl > 10*time.Second {
		t.Fatalf("existing TTL was extended to %v", ttl)
	}
	if v, err := DecrByClient(ctx, cli, key, time.Minute); err != nil || v != 1 {
		t.Fatalf("decr = %d, %v", v, err)
	}
}

func TestIncrWithoutExpireKeepsKey(t *testing.T) {
	cli := testRedis(t)
	ctx := t.Context()
	key := "kit:test:incr:" + t.Name()
	t.Cleanup(func() { cli.Del(context.Background(), key) })

	if _, err := DecrByClient(ctx, cli, key, 0); err != nil {
		t.Fatal(err)
	}
	if v := cli.Get(ctx, key).Val(); v != "-1" {
		t.Fatalf("value = %q, want -1 kept without TTL", v)
	}
	if ttl := cli.TTL(ctx, key).Val(); ttl != -1 {
		t.Fatalf("ttl = %v, want none", ttl)
	}
}
