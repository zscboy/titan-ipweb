package utils

import (
	"time"
	"titan-ipweb/internal/types"
)

func Success(data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code: 0,
		Msg:  "",
		Time: time.Now().Unix(),
		Data: data,
	}
}

func Error(err error) *types.BaseResponse {
	return &types.BaseResponse{
		Code: -1,
		Msg:  err.Error(),
		Time: time.Now().Unix(),
		Data: nil,
	}
}
