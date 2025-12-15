package model

import (
	"context"
	"errors"
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
	DownloadRateLimit int64  `redis:"download_rate_limit"`
	MaxBandwidthLimit int64  `redis:"max_bandwidth_limit"`
	TotalTrafficLimit int64  `redis:"total_traffic_limit"`
	CreateTime        int64  `redis:"create_time"`
	DeprecatedTime    int64  `redis:"deprecated_time"`
	Status            string `redis:"status"`
	StartTime         int64  `redis:"start_time"`
	EndTime           int64  `redis:"end_time"`
	UserID            string `redis:"user_id"`
	PopID             string `redis:"pop_id"`
}

func subUserKey(username string) string {
	return fmt.Sprintf(redisKeySubUserTable, username)
}

func subUserListKey(uuid string) string {
	return fmt.Sprintf(redisKeyUserSubUserZset, uuid)
}

func deprecatedSubUserListKey(uuid string) string {
	return fmt.Sprintf(redisKeyInvalidSubUserZset, uuid)
}

func SaveSubUser(rdb *redis.Redis, subUser *SubUser) error {
	if subUser == nil {
		return fmt.Errorf("subUser is nil")
	}

	if subUser.Username == "" {
		return fmt.Errorf("empty Username")
	}

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

	key = subUserListKey(uuid)
	_, err = rdb.Zrem(key, subUsername)
	if err != nil {
		return err
	}

	key = deprecatedSubUserListKey(uuid)
	_, err = rdb.Zrem(key, subUsername)
	return err
}

func GetSubUser(rdb *redis.Redis, username string) (*SubUser, error) {
	key := subUserKey(username)
	data, err := rdb.Hgetall(key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	subUser := &SubUser{}
	if err := mapToStruct(data, subUser); err != nil {
		return nil, err
	}
	return subUser, nil
}

func AddSubUserToList(rdb *redis.Redis, uuid string, subUsername string) error {
	key := subUserListKey(uuid)
	_, err := rdb.Zadd(key, time.Now().Unix(), subUsername)
	return err
}

func GetSubUsers(ctx context.Context, rdb *redis.Redis, uuid string, start, stop int) ([]*SubUser, error) {
	key := subUserListKey(uuid)
	usernames, err := rdb.Zrange(key, int64(start), int64(stop))
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

func AddSubUserToDeprecatedList(rdb *redis.Redis, uuid string, subUsername string) error {
	key := deprecatedSubUserListKey(uuid)
	_, err := rdb.Zadd(key, time.Now().Unix(), subUsername)
	if err != nil {
		return err
	}

	key = subUserListKey(uuid)
	_, err = rdb.Zrem(key, subUsername)
	return err

}

// TODO: split by start and stop
func GetDeprecatedSubUsers(ctx context.Context, rdb *redis.Redis, uuid string, start, end int) ([]*SubUser, error) {
	key := deprecatedSubUserListKey(uuid)
	usernames, err := rdb.Zrange(key, int64(start), int64(end))
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

func SubUserCount(rdb *redis.Redis, uuid string) (int, error) {
	key := subUserListKey(uuid)
	return rdb.Zcard(key)
}

func DeprecatedSubUserCount(rdb *redis.Redis, uuid string) (int, error) {
	key := deprecatedSubUserListKey(uuid)
	return rdb.Zcard(key)
}
