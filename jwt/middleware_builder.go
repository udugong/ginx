package jwt

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// MiddlewareBuilder 创建一个校验登录的 middleware
// ignorePath: 默认使用 func(*gin.Context) bool { return false } 也就是全部不忽略.
type MiddlewareBuilder[T any] struct {
	ignorePath func(*gin.Context) bool // Middleware 方法中忽略认证的路径
	manager    *Management[T]
	nowFunc    func() time.Time // 控制 jwt 的时间
}

func newMiddlewareBuilder[T any](m *Management[T]) *MiddlewareBuilder[T] {
	return &MiddlewareBuilder[T]{
		manager: m,
		ignorePath: func(*gin.Context) bool {
			return false
		},
		nowFunc: m.nowFunc,
	}
}

// IgnorePathFunc 设置忽略资源令牌认证的路径.
func (m *MiddlewareBuilder[T]) IgnorePathFunc(fn func(*gin.Context) bool) *MiddlewareBuilder[T] {
	m.ignorePath = fn
	return m
}

func (m *MiddlewareBuilder[T]) IgnorePath(paths ...string) *MiddlewareBuilder[T] {
	return m.IgnorePathFunc(staticIgnorePaths(paths...))
}

// IgnoreFullPath 忽略匹配的完整路径.
// 例如: "/user/:id"
func (m *MiddlewareBuilder[T]) IgnoreFullPath(fullPaths ...string) *MiddlewareBuilder[T] {
	return m.IgnorePathFunc(staticIgnoreFullPaths(fullPaths...))
}

func (m *MiddlewareBuilder[T]) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不需要校验
		if m.ignorePath(c) {
			return
		}

		// 提取 token
		tokenStr := m.manager.extractTokenString(c)
		if tokenStr == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 校验 token
		clm, err := m.manager.VerifyAccessToken(tokenStr,
			jwt.WithTimeFunc(m.nowFunc))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置 claims
		m.manager.SetClaims(c, clm)
	}
}

// staticIgnorePaths 设置静态忽略的路径.
func staticIgnorePaths(paths ...string) func(c *gin.Context) bool {
	s := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		s[path] = struct{}{}
	}
	return func(c *gin.Context) bool {
		_, ok := s[c.Request.URL.Path]
		return ok
	}
}

// staticIgnoreFullPaths 设置静态忽略完整路径.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func staticIgnoreFullPaths(paths ...string) func(c *gin.Context) bool {
	s := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		s[path] = struct{}{}
	}
	return func(c *gin.Context) bool {
		_, ok := s[c.FullPath()]
		return ok
	}
}
