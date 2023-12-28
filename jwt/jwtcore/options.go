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

func WithDecryptKey[T jwt.Claims, PT Claims[T]](key string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.DecryptKey = key
	})
}

func WithMethod[T jwt.Claims, PT Claims[T]](method jwt.SigningMethod) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.Method = method
	})
}

func WithIssuer[T jwt.Claims, PT Claims[T]](issuer string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.Issuer = issuer
	})
}

func WithGenIDFunc[T jwt.Claims, PT Claims[T]](fn func() string) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.genIDFn = fn
	})
}

func WithTimeFunc[T jwt.Claims, PT Claims[T]](fn func() time.Time) Option[T, PT] {
	return optionFunc[T, PT](func(t *TokenManagerServer[T, PT]) {
		t.timeFunc = fn
	})
}
