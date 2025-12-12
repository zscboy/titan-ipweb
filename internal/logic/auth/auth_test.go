package auth

import (
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	secret := "c96ce150-d1ab-11f0-adbd-e30c81911f62"
	uuid := "d1h50rddpgj9uqcrdesg"
	email := "yuanstar00@gmail.com"
	token, err := generateToken(secret, uuid, email, time.Second*600)
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	t.Logf("token:%s", token)

}
