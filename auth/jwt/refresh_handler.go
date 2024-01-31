package jwt

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/udugong/token"
)

// RefreshManager 定义刷新令牌管理器.
type RefreshManager[T jwt.Claims] struct {
	// accessTM 资源令牌管理.
	accessTM token.Manager[T]

	// refreshTM 刷新令牌管理.
	refreshTM token.Manager[T]

	// rotateRefreshToken 是否轮换刷新令牌.
	// 默认为 false.
	rotateRefreshToken bool

	// refreshAuthHandler 认证的处理函数.
	// 默认使用 refreshTM 作为参数创建的 MiddlewareBuilder 进行认证.
	refreshAuthHandler gin.HandlerFunc

	// getClaims 获取 Claims.
	// 如果更改了 refreshAuthHandler 中设置 Claims 的方法,则需要匹配 setClaims 来获取.
	getClaims func(*gin.Context) (T, bool)

	// accessTokenSetterFn 资源令牌设置函数.
	// 默认把 access token 设置到 key="x-access-token" 的请求头中.
	accessTokenSetterFn TokenSetterFunc

	// refreshTokenSetterFn 刷新令牌设置函数.
	// 默认把 refresh token 设置到 key="x-refresh-token" 的请求头中.
	refreshTokenSetterFn TokenSetterFunc

	// responseHandler 响应函数.
	// 默认返回 HTTP 响应码为 204 的响应.
	responseHandler gin.HandlerFunc
}

// TokenSetterFunc 令牌设置函数.
// 设置 token 到 gin.Context 中.
type TokenSetterFunc func(c *gin.Context, token string)

// NewRefreshManager 创建一个刷新令牌管理器.
func NewRefreshManager[T jwt.Claims](
	accessTM token.Manager[T], refreshTM token.Manager[T],
	options ...Option[T]) *RefreshManager[T] {
	m := &RefreshManager[T]{
		accessTM:           accessTM,
		refreshTM:          refreshTM,
		rotateRefreshToken: false,
	}
	m.refreshAuthHandler = NewMiddlewareBuilder[T](refreshTM).Build()
	m.getClaims = func(c *gin.Context) (T, bool) {
		return ClaimsFromContext[T](c.Request.Context())
	}
	m.accessTokenSetterFn = func(c *gin.Context, token string) {
		c.Header("x-access-token", token)
	}
	m.refreshTokenSetterFn = func(c *gin.Context, token string) {
		c.Header("x-refresh-token", token)
	}
	m.responseHandler = func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	}
	return m.WithOptions(options...)
}

type Option[T jwt.Claims] interface {
	apply(*RefreshManager[T])
}

type optionFunc[T jwt.Claims] func(*RefreshManager[T])

func (f optionFunc[T]) apply(m *RefreshManager[T]) {
	f(m)
}

// WithRotateRefreshToken 是否轮换 refresh token.
func WithRotateRefreshToken[T jwt.Claims](isRotate bool) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.rotateRefreshToken = isRotate
	})
}

// WithRefreshAuthHandler 更改刷新令牌函数的验证 gin.HandlerFunc.
// 请小心使用该方法.
func WithRefreshAuthHandler[T jwt.Claims](fn gin.HandlerFunc) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.refreshAuthHandler = fn
	})
}

// WithGetClaims 更改获取 Claims 方式.
// 使用该方法需要与 refreshAuthHandler 中设置 Claims 的方式匹配.
func WithGetClaims[T jwt.Claims](fn func(*gin.Context) (T, bool)) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.getClaims = fn
	})
}

// WithAccessTokenSetter 更改设置 access token 的方式.
func WithAccessTokenSetter[T jwt.Claims](fn TokenSetterFunc) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.accessTokenSetterFn = fn
	})
}

// WithRefreshTokenSetter 更改设置 refresh token 的方式
func WithRefreshTokenSetter[T jwt.Claims](fn TokenSetterFunc) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.refreshTokenSetterFn = fn
	})
}

// WithResponseSetter 更改刷新令牌函数的响应.
func WithResponseSetter[T jwt.Claims](fn gin.HandlerFunc) Option[T] {
	return optionFunc[T](func(m *RefreshManager[T]) {
		m.responseHandler = fn
	})
}

// Handler 刷新令牌的 gin.HandlerFunc.
func (m *RefreshManager[T]) Handler(c *gin.Context) {
	m.refreshAuthHandler(c)
	if c.IsAborted() {
		return
	}
	clm, ok := m.getClaims(c)
	if !ok {
		// 不应该命中该分支
		c.Status(http.StatusInternalServerError)
		log.Println("未知错误")
		return
	}

	accessToken, err := m.accessTM.GenerateToken(clm)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Printf("生成 access token 失败; err: %v", err)
		return
	}

	m.accessTokenSetterFn(c, accessToken)

	// 轮换刷新令牌
	if m.rotateRefreshToken {
		refreshToken, err := m.refreshTM.GenerateToken(clm)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("生成 refresh token 失败; err: %v", err)
			return
		}
		m.refreshTokenSetterFn(c, refreshToken)
	}
	m.responseHandler(c)
}

func (m *RefreshManager[T]) WithOptions(opts ...Option[T]) *RefreshManager[T] {
	c := m.clone()
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func (m *RefreshManager[T]) clone() *RefreshManager[T] {
	copyHandler := *m
	return &copyHandler
}
