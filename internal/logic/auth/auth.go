package auth

import (
	"time"
	"titan-ipweb/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
)

func generateToken(accessSecret, uuid, email string, expire time.Duration) (string, error) {
	claims := middleware.Claims{
		UserId: uuid,
		Email:  email,
		//SessionKey: sessionKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
	}

	// 使用 HS256 算法生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}
