package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthCtxValue struct {
	UserId    string
	Email     string
	Role      string
	ExpiresAt int64
}

type AuthCtxKey string

const AuthKey AuthCtxKey = "auth"

type TokenType string

// JWT Claims 结构体
type Claims struct {
	UserId     string    `json:"user_id"`
	SessionKey int64     `json:"sskey"`
	Email      string    `json:"email"`
	Type       TokenType `json:"token_type"`

	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	AccessSecretKey string
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		AccessSecretKey: secretKey,
	}
}

// 用于解析 JWT Token 并验证有效性
func (m *AuthMiddleware) parseToken(tokenString string) (AuthCtxValue, error) {
	// 解析并验证 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.AccessSecretKey), nil
	})

	if err != nil {
		return AuthCtxValue{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return AuthCtxValue{}, fmt.Errorf("invalid token")
	}

	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		return AuthCtxValue{}, fmt.Errorf("token expired")
	}

	return AuthCtxValue{UserId: claims.UserId, Email: claims.Email}, nil
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// 从 Authorization 头中获取 token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		// 格式应为 "Bearer <token>"
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// 提取 token 部分
		tokenString := parts[1]

		// 解析并验证 token
		authValue, err := m.parseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, AuthKey, authValue)
		r = r.WithContext(ctx)
		next(w, r)
	}
}
