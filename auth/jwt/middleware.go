package jwt

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/udugong/token"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "Bearer"
)

// MiddlewareBuilder 定义认证的中间件构建器.
type MiddlewareBuilder[T jwt.Claims] struct {
	// Middleware 中忽略认证路径的方法.
	// 默认使用 func(*gin.Context) bool { return false } 也就是全部不忽略.
	ignorePath func(*gin.Context) bool

	// Middleware 中提取 token 字符串的方法.
	// 默认从 authorization 请求头中,获取按 Bearer 分割获取.
	extractToken func(*gin.Context) string

	// Middleware 中设置 Claims 的方法.
	// 默认设置到 key=claimsKey{} 的 context.Context 中.
	// 通过 ClaimsFromContext[T]() 获取 Claims.
	setClaims func(*gin.Context, T)

	TokenManager token.Manager[T]
}

// NewMiddlewareBuilder 创建一个认证的中间件构建器.
func NewMiddlewareBuilder[T jwt.Claims](
	m token.Manager[T]) *MiddlewareBuilder[T] {
	return &MiddlewareBuilder[T]{
		ignorePath: func(*gin.Context) bool {
			return false
		},
		extractToken: extractToken,
		setClaims: func(c *gin.Context, t T) {
			c.Request = c.Request.WithContext(
				ContextWithClaims(c.Request.Context(), t))
		},
		TokenManager: m,
	}
}

// IgnorePathFunc 设置忽略认证路径.
func (m *MiddlewareBuilder[T]) IgnorePathFunc(fn func(*gin.Context) bool) *MiddlewareBuilder[T] {
	m.ignorePath = fn
	return m
}

// SetExtractTokenFunc 设置提取 token 字符串的方法.
func (m *MiddlewareBuilder[T]) SetExtractTokenFunc(fn func(*gin.Context) string) *MiddlewareBuilder[T] {
	m.extractToken = fn
	return m
}

// SetClaimsFunc 设置 Claims 的方法.
func (m *MiddlewareBuilder[T]) SetClaimsFunc(fn func(*gin.Context, T)) *MiddlewareBuilder[T] {
	m.setClaims = fn
	return m
}

// IgnoreFullPath 忽略匹配的完整路径.
// 例如: "/user/:id"
func (m *MiddlewareBuilder[T]) IgnoreFullPath(fullPaths ...string) *MiddlewareBuilder[T] {
	return m.IgnorePathFunc(ignoreFullPaths(fullPaths...))
}

// Build 构建认证中间件.
func (m *MiddlewareBuilder[T]) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不需要校验
		if m.ignorePath(c) {
			return
		}

		// 提取 token
		tokenStr := m.extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 校验 token
		clm, err := m.TokenManager.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置 claims
		m.setClaims(c, clm)
	}
}

// claimsKey 定义从 context.Context 中设置/获取 claims 的 key.
type claimsKey struct{}

// ContextWithClaims 为 claims 创建 context.
func ContextWithClaims[T jwt.Claims](ctx context.Context, claims T) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// ClaimsFromContext 从 context 中获取 claims.
// 如果没有正确的 claims 则返回 false.
func ClaimsFromContext[T jwt.Claims](ctx context.Context) (T, bool) {
	var zeroClm T
	v := ctx.Value(claimsKey{})
	clm, ok := v.(T)
	if !ok {
		return zeroClm, false
	}
	return clm, true
}

// ignoreFullPaths 设置忽略完整路径.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func ignoreFullPaths(paths ...string) func(c *gin.Context) bool {
	s := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		s[path] = struct{}{}
	}
	return func(c *gin.Context) bool {
		_, ok := s[c.FullPath()]
		return ok
	}
}

// extractToken 提取 token 字符串.
func extractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader(authorizationHeader)
	if authCode == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(bearerPrefix) + 1)
	b.WriteString(bearerPrefix)
	b.WriteString(" ")
	prefix := b.String()
	if strings.HasPrefix(authCode, prefix) {
		return authCode[len(prefix):]
	}
	return ""
}
