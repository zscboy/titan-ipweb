package model

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type SubUser struct {
	Username          string `redis:"username"`
	Password          string `redis:"password"`
	ServerAddress     string `redis:"server_address"`
	UploadRateLimit   int64  `redis:"upload_rate_limit"`
	DownloadRateLImit int64  `redis:"download_rate_limit"`
	TotalTrafficLimit int64  `redis:"total_traffic_limit"`
}

func subUserKey(username string) string {
	return fmt.Sprintf(redisKeySubUserTable, username)
}

func userSubUserKey(uuid string) string {
	return fmt.Sprintf(redisKeyUserSubUserZset, uuid)
}

func SaveSubUser(rdb *redis.Redis, subUser *SubUser) error {
	m, err := structToMap(subUser)
	if err != nil {
		return err
	}

	key := subUserKey(subUser.Username)
	return rdb.Hmset(key, m)
}

func RemoveSubUser(rdb *redis.Redis, uuid, subUsername string) error {
	key := subUserKey(subUsername)
	_, err := rdb.Del(key)
	if err != nil {
		return err
	}

	key = userSubUserKey(uuid)
	_, err = rdb.Zrem(key, subUsername)
	return err
}

func GetSubUser(rdb *redis.Redis, username string) (*SubUser, error) {
	key := subUserKey(username)
	data, err := rdb.Hgetall(key)
	if err != nil {
		return nil, err
	}

	subUser := &SubUser{}
	if err := mapToStruct(data, subUser); err != nil {
		return nil, err
	}
	return subUser, nil
}

func AddSubUser(rdb *redis.Redis, uuid string, subUsername string) error {
	key := userSubUserKey(uuid)
	_, err := rdb.Zadd(key, time.Now().Unix(), subUsername)
	return err
}

func GetUserSubUsers(ctx context.Context, rdb *redis.Redis, uuid string) ([]*SubUser, error) {
	key := userSubUserKey(uuid)
	usernames, err := rdb.Zrange(key, 0, -1)
	if err != nil {
		return nil, err
	}

	pipe, err := rdb.TxPipeline()
	if err != nil {
		return nil, err
	}

	for _, username := range usernames {
		key := subUserKey(username)
		pipe.HGetAll(ctx, key)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	subUsers := make([]*SubUser, 0, len(cmds))
	for _, cmd := range cmds {
		result, err := cmd.(*goredis.MapStringStringCmd).Result()
		if err != nil {
			logx.Errorf("ListNode parse result failed:%s", err.Error())
			continue
		}

		subUser := SubUser{}
		err = mapToStruct(result, &subUser)
		if err != nil {
			logx.Errorf("ListNode mapToStruct error:%s", err.Error())
			continue
		}

		subUsers = append(subUsers, &subUser)
	}

	return subUsers, nil

}
