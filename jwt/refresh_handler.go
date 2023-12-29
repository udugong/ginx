package jwt

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/udugong/ginx/jwt/jwtcore"
)

// RefreshManager 定义刷新令牌管理器.
type RefreshManager[T jwt.Claims, PT jwtcore.Claims[T]] struct {
	// accessTM 资源令牌管理.
	accessTM jwtcore.TokenManager[T, PT]

	// refreshTM 刷新令牌管理.
	refreshTM jwtcore.TokenManager[T, PT]

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
func NewRefreshManager[T jwt.Claims, PT jwtcore.Claims[T]](
	accessTM jwtcore.TokenManager[T, PT], refreshTM jwtcore.TokenManager[T, PT],
	options ...Option[T, PT]) *RefreshManager[T, PT] {
	m := &RefreshManager[T, PT]{
		accessTM:           accessTM,
		refreshTM:          refreshTM,
		rotateRefreshToken: false,
	}
	m.refreshAuthHandler = NewMiddlewareBuilder[T, PT](refreshTM).Build()
	m.getClaims = func(c *gin.Context) (T, bool) {
		v, ok := c.Get(claimsKey)
		clm, ok := v.(T)
		return clm, ok
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

type Option[T jwt.Claims, PT jwtcore.Claims[T]] interface {
	apply(*RefreshManager[T, PT])
}

type optionFunc[T jwt.Claims, PT jwtcore.Claims[T]] func(*RefreshManager[T, PT])

func (f optionFunc[T, PT]) apply(m *RefreshManager[T, PT]) {
	f(m)
}

func WithRotateRefreshToken[T jwt.Claims, PT jwtcore.Claims[T]](isRotate bool) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.rotateRefreshToken = isRotate
	})
}

func WithRefreshAuthHandler[T jwt.Claims, PT jwtcore.Claims[T]](fn gin.HandlerFunc) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.refreshAuthHandler = fn
	})
}

func WithGetClaims[T jwt.Claims, PT jwtcore.Claims[T]](fn func(*gin.Context) (T, bool)) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.getClaims = fn
	})
}

func WithAccessTokenSetter[T jwt.Claims, PT jwtcore.Claims[T]](fn TokenSetterFunc) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.accessTokenSetterFn = fn
	})
}

func WithRefreshTokenSetter[T jwt.Claims, PT jwtcore.Claims[T]](fn TokenSetterFunc) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.refreshTokenSetterFn = fn
	})
}

func WithResponseSetter[T jwt.Claims, PT jwtcore.Claims[T]](fn gin.HandlerFunc) Option[T, PT] {
	return optionFunc[T, PT](func(m *RefreshManager[T, PT]) {
		m.responseHandler = fn
	})
}

// Handler 刷新令牌的 gin.HandlerFunc.
func (m *RefreshManager[T, PT]) Handler(c *gin.Context) {
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

func (m *RefreshManager[T, PT]) WithOptions(opts ...Option[T, PT]) *RefreshManager[T, PT] {
	c := m.clone()
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func (m *RefreshManager[T, PT]) clone() *RefreshManager[T, PT] {
	copyHandler := *m
	return &copyHandler
}
