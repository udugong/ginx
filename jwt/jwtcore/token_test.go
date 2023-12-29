package jwtcore

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type MyClaims struct {
	Uid int64 `json:"uid,omitempty"`
	RegisteredClaims
}

func TestNewTokenHandler(t *testing.T) {
	var genIDFn func() string
	var timeFn func() time.Time
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name          string
		expire        time.Duration
		encryptionKey string
		want          *TokenManagerServer[T, PT]
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name:          "normal",
			expire:        defaultExpire,
			encryptionKey: encryptionKey,
			want: &TokenManagerServer[MyClaims, *MyClaims]{
				Expire:        defaultExpire,
				EncryptionKey: encryptionKey,
				DecryptKey:    encryptionKey,
				Method:        jwt.SigningMethodHS256,
				genIDFn:       genIDFn,
				timeFunc:      timeFn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTokenManagerServer[MyClaims, *MyClaims](tt.encryptionKey, tt.expire)
			got.genIDFn = genIDFn
			got.timeFunc = timeFn
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenHandler_GenerateToken(t *testing.T) {
	h := defaultHandler
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name    string
		clm     T
		want    string
		wantErr error
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name:    "normal",
			clm:     defaultClaims,
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.B9sIBtCtX5kp8pk0fjpcy-8HVa991qU5L5nles7Nblw",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.GenerateToken(tt.clm)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenHandler_VerifyToken(t *testing.T) {
	type testCase[T jwt.Claims, PT Claims[T]] struct {
		name    string
		h       *TokenManagerServer[T, PT]
		token   string
		want    T
		wantErr error
	}
	tests := []testCase[MyClaims, *MyClaims]{
		{
			name:    "normal",
			h:       defaultHandler,
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.B9sIBtCtX5kp8pk0fjpcy-8HVa991qU5L5nles7Nblw",
			want:    defaultClaims,
			wantErr: nil,
		},
		{
			// token 过期了
			name: "token_expired",
			h: NewTokenManagerServer[MyClaims, *MyClaims](encryptionKey, defaultExpire,
				WithTimeFunc[MyClaims, *MyClaims](func() time.Time {
					return time.UnixMilli(1695671200000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.B9sIBtCtX5kp8pk0fjpcy-8HVa991qU5L5nles7Nblw",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			// token 签名错误
			name:  "bad_sign_key",
			h:     defaultHandler,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.jnzq7EJftxHk82jxl645w875Z0C8yn9WG3uGKhQuLm4",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			// 错误的 token
			name:  "bad_token",
			h:     defaultHandler,
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.VerifyToken(tt.token)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

var (
	encryptionKey = "sign key"
	nowTime       = time.UnixMilli(1695571200000)
	defaultExpire = 10 * time.Minute
	defaultClaims = MyClaims{
		Uid: 1,
		RegisteredClaims: RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(defaultExpire)),
			IssuedAt:  jwt.NewNumericDate(nowTime),
		},
	}
	defaultHandler = NewTokenManagerServer[MyClaims, *MyClaims](
		encryptionKey, defaultExpire,
		WithTimeFunc[MyClaims, *MyClaims](func() time.Time {
			return nowTime
		}),
	)
)
