package redisx

import "github.com/redis/go-redis/v9"

var GlobalClient redis.UniversalClient

func SetClient(cli redis.UniversalClient) {
	GlobalClient = cli
}
