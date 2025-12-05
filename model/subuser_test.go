package model

import (
	"testing"

	"github.com/mitchellh/mapstructure"
)

func TestSubUser(t *testing.T) {
	subUser := SubUser{}
	data := make(map[string]interface{})
	if err := mapstructure.Decode(subUser, &data); err != nil {
		t.Logf("decode failed:%v", err)
		return
	}

	t.Logf("subUser:%v", data)

}
