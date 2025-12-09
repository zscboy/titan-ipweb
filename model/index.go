package model

import "github.com/zeromicro/go-zero/core/stores/redis"

func UserIndex(rdb *redis.Redis) (int64, error) {
	return rdb.Incr(redisKeyUserIndex)
}
