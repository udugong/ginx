package jwt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/udugong/ginx/jwt/jwtcore"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "Bearer"
	// claimsKey 定义了存储 Claims 在 gin.Context 中的 key.
	claimsKey = "claims"
)

// MiddlewareBuilder 定义认证的中间件构建器.
type MiddlewareBuilder[T jwt.Claims, PT jwtcore.Claims[T]] struct {
	// Middleware 中忽略认证路径的方法.
	// 默认使用 func(*gin.Context) bool { return false } 也就是全部不忽略.
	ignorePath func(*gin.Context) bool

	// Middleware 中提取 token 字符串的方法.
	// 默认从 authorization 请求头中,获取按 Bearer 分割获取.
	extractToken func(*gin.Context) string

	// Middleware 中设置 Claims 的方法.
	// 默认设置到 key="claims" 的 gin.Context 中.
	setClaims func(*gin.Context, T)
	jwtcore.TokenManager[T, PT]
}

// NewMiddlewareBuilder 创建一个认证的中间件构建器.
func NewMiddlewareBuilder[T jwt.Claims, PT jwtcore.Claims[T]](
	m jwtcore.TokenManager[T, PT]) *MiddlewareBuilder[T, PT] {
	return &MiddlewareBuilder[T, PT]{
		ignorePath: func(*gin.Context) bool {
			return false
		},
		extractToken: extractToken,
		setClaims: func(c *gin.Context, t T) {
			c.Set(claimsKey, t)
		},
		TokenManager: m,
	}
}

// IgnorePathFunc 设置忽略认证路径.
func (m *MiddlewareBuilder[T, PT]) IgnorePathFunc(fn func(*gin.Context) bool) *MiddlewareBuilder[T, PT] {
	m.ignorePath = fn
	return m
}

// SetExtractTokenFunc 设置提取 token 字符串的方法.
func (m *MiddlewareBuilder[T, PT]) SetExtractTokenFunc(fn func(*gin.Context) string) *MiddlewareBuilder[T, PT] {
	m.extractToken = fn
	return m
}

// SetClaimsFunc 设置 Claims 的方法.
func (m *MiddlewareBuilder[T, PT]) SetClaimsFunc(fn func(*gin.Context, T)) *MiddlewareBuilder[T, PT] {
	m.setClaims = fn
	return m
}

// IgnoreFullPath 忽略匹配的完整路径.
// 例如: "/user/:id"
func (m *MiddlewareBuilder[T, PT]) IgnoreFullPath(fullPaths ...string) *MiddlewareBuilder[T, PT] {
	return m.IgnorePathFunc(ignoreFullPaths(fullPaths...))
}

// Build 构建认证中间件.
func (m *MiddlewareBuilder[T, PT]) Build() gin.HandlerFunc {
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
	b.WriteString(bearerPrefix)
	b.WriteString(" ")
	prefix := b.String()
	if strings.HasPrefix(authCode, prefix) {
		return authCode[len(prefix):]
	}
	return ""
}
