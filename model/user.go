package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type User struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
}

func HSetUser(rdb *redis.Redis, user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if user.UUID == "" {
		return fmt.Errorf("empty uuid")
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return rdb.Hset(redisKeyUserTable, user.UUID, string(data))

}

func HGetUser(rdb *redis.Redis, uuid string) (*User, error) {
	if uuid == "" {
		return nil, fmt.Errorf("empty uuid")
	}

	val, err := rdb.Hget(redisKeyUserTable, uuid)
	if errors.Is(err, redis.Nil) { // key not found
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
