package jwtcore

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager jwt token 的管理接口.
type TokenManager[T jwt.Claims, PT Claims[T]] interface {
	GenerateToken(clm T) (string, error)
	VerifyToken(token string, opts ...jwt.ParserOption) (T, error)
}

// TokenManagerServer 定义 jwt token 的处理程序.
type TokenManagerServer[T jwt.Claims, PT Claims[T]] struct {
	EncryptionKey string            // 加密密钥
	DecryptKey    string            // 解密密钥
	Method        jwt.SigningMethod // 签名方式
	Expire        time.Duration     // 有效期
	Issuer        string            // 签发人
	genIDFn       func() string     // 生成 JWT ID (jti) 的函数
	timeFunc      func() time.Time  // 控制 jwt 的时间
}

// NewTokenManagerServer 创建一个 jwt token 的处理服务.
// Method: 默认使用 jwt.SigningMethodHS256 对称签名方式.
// DecryptKey: 默认与 EncryptionKey 相同.
func NewTokenManagerServer[T jwt.Claims, PT Claims[T]](encryptionKey string,
	expire time.Duration, options ...Option[T, PT]) *TokenManagerServer[T, PT] {
	manager := &TokenManagerServer[T, PT]{
		Expire:        expire,
		EncryptionKey: encryptionKey,
		DecryptKey:    encryptionKey,
		Method:        jwt.SigningMethodHS256,
		genIDFn: func() string {
			return ""
		},
		timeFunc: time.Now,
	}
	return manager.WithOptions(options...)
}

// GenerateToken 生成一个 jwt token.
func (t *TokenManagerServer[T, PT]) GenerateToken(clm T) (string, error) {
	nowTime := t.timeFunc()
	p := PT(&clm)
	p.SetIssuer(t.Issuer)
	p.SetExpiresAt(jwt.NewNumericDate(nowTime.Add(t.Expire)))
	p.SetIssuedAt(jwt.NewNumericDate(nowTime))
	p.SetID(t.genIDFn())
	token := jwt.NewWithClaims(t.Method, clm)
	return token.SignedString([]byte(t.EncryptionKey))
}

// VerifyToken 认证 token 并返回 claims 与 error.
func (t *TokenManagerServer[T, PT]) VerifyToken(token string, opts ...jwt.ParserOption) (T, error) {
	var zeroClm T
	clm := zeroClm
	var clmPtr any = &clm
	opts = append(opts, jwt.WithTimeFunc(t.timeFunc))
	withClaims, err := jwt.ParseWithClaims(token, clmPtr.(jwt.Claims),
		func(*jwt.Token) (interface{}, error) {
			return []byte(t.DecryptKey), nil
		},
		opts...,
	)
	if err != nil || !withClaims.Valid {
		return zeroClm, fmt.Errorf("验证失败: %v", err)
	}
	return clm, nil
}

func (t *TokenManagerServer[T, PT]) WithOptions(opts ...Option[T, PT]) *TokenManagerServer[T, PT] {
	c := t.clone()
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func (t *TokenManagerServer[T, PT]) clone() *TokenManagerServer[T, PT] {
	copyHandler := *t
	return &copyHandler
}
