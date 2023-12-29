package jwtcore

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// An Option configures a TokenManagerServer.
type Option[T jwt.Claims, PT Claims[T]] interface {
	apply(*TokenManagerServer[T, PT])
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc[T jwt.Claims, PT Claims[T]] func(*TokenManagerServer[T, PT])

func (f optionFunc[T, PT]) apply(manager *TokenManagerServer[T, PT]) {
	f(manager)
}

// WithDecryptKey 设置解密密钥.
func WithDecryptKey[T jwt.Claims, PT Claims[T]](key string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.DecryptKey = key
	})
}

// WithMethod 设置 jwt 的签名方式.
func WithMethod[T jwt.Claims, PT Claims[T]](method jwt.SigningMethod) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.Method = method
	})
}

// WithIssuer 设置签发者.
func WithIssuer[T jwt.Claims, PT Claims[T]](issuer string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.Issuer = issuer
	})
}

// WithGenIDFunc 设置生成 jwt ID 的函数.
func WithGenIDFunc[T jwt.Claims, PT Claims[T]](fn func() string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.genIDFn = fn
	})
}

// WithTimeFunc 设置时间函数.
// 可以固定 jwt 的时间.
func WithTimeFunc[T jwt.Claims, PT Claims[T]](fn func() time.Time) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.timeFunc = fn
	})
}
