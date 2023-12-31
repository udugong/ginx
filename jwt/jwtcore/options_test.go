package jwtcore

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithDecryptKey(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want string
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name: "normal",
			fn:   withNop[MyClaims, *MyClaims],
			want: encryptionKey,
		},
		{
			name: "set_another_key",
			fn: func() Option[MyClaims, *MyClaims] {
				return WithDecryptKey[MyClaims, *MyClaims]("another sign key")
			},
			want: "another sign key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](
				encryptionKey, defaultExpire, tt.fn()).DecryptKey
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithGenIDFunc(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want string
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name: "normal",
			fn:   withNop[MyClaims, *MyClaims],
			want: "",
		},
		{
			name: "set_another_gen_id_func",
			fn: func() Option[MyClaims, *MyClaims] {
				return WithGenIDFunc[MyClaims, *MyClaims](func() string {
					return "unique id"
				})
			},
			want: "unique id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](
				encryptionKey, defaultExpire, tt.fn()).genIDFn()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithIssuer(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want string
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name: "normal",
			fn:   withNop[MyClaims, *MyClaims],
			want: "",
		},
		{
			name: "set_another_issuer",
			fn: func() Option[MyClaims, *MyClaims] {
				return WithIssuer[MyClaims, *MyClaims]("foo")
			},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](
				encryptionKey, defaultExpire, tt.fn()).Issuer
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithMethod(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want jwt.SigningMethod
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name: "normal",
			fn:   withNop[MyClaims, *MyClaims],
			want: jwt.SigningMethodHS256,
		},
		{
			name: "set_another_method",
			fn: func() Option[MyClaims, *MyClaims] {
				return WithMethod[MyClaims, *MyClaims](jwt.SigningMethodHS384)
			},
			want: jwt.SigningMethodHS384,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](
				encryptionKey, defaultExpire, tt.fn()).Method
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithTimeFunc(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want int64
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name: "set_default_time_func",
			fn: func() Option[MyClaims, *MyClaims] {
				return WithTimeFunc[MyClaims, *MyClaims](func() time.Time {
					return nowTime
				})
			},
			want: 1695571200000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](
				encryptionKey, defaultExpire, tt.fn()).timeFunc().UnixMilli()
			assert.Equal(t, tt.want, got)
		})
	}
}

func withNop[T jwt.Claims, PT Claims[T]]() Option[T, PT] {
	return optionFunc[T, PT](func(m *TokenManagerServer[T, PT]) {})
}
