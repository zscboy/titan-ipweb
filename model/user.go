package model

import (
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type User struct {
	UUID                  string `redis:"uuid"`
	Email                 string `redis:"email"`
	Index                 int64  `redis:"index"`
	MaxBandwidthLimit     int64  `redis:"max_bandwidth_limit"`
	MaxBandwidthAllocated int64  `redis:"max_bandwidth_allocated"`
	TotalTrafficLimit     int64  `redis:"total_traffic_limit"`
	TotalTrafficAllocated int64  `redis:"total_traffic_allocated"`
}

func userKey(uuid string) string {
	return fmt.Sprintf(redisKeyUserTable, uuid)
}

func SaveUser(rdb *redis.Redis, user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if user.UUID == "" {
		return fmt.Errorf("empty uuid")
	}

	m, err := structToMap(user)
	if err != nil {
		return err
	}

	key := userKey(user.UUID)
	return rdb.Hmset(key, m)
}

func GetUser(rdb *redis.Redis, uuid string) (*User, error) {
	if uuid == "" {
		return nil, fmt.Errorf("empty uuid")
	}

	key := userKey(uuid)

	data, err := rdb.Hgetall(key)
	if errors.Is(err, redis.Nil) { // key not found
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	user := &User{}
	if err := mapToStruct(data, user); err != nil {
		return nil, err
	}

	return user, nil
}

// limit user length
func GetAllSubUsername(rdb *redis.Redis, uuid string) ([]string, error) {
	key := subUserListKey(uuid)
	usernames, err := rdb.Zrange(key, 0, -1)
	if err != nil {
		return nil, err
	}
	return usernames, nil
}
