package jwt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/udugong/ginx/jwt/jwtcore"
)

type Claims struct {
	Uid int64 `json:"uid,omitempty"`
	jwtcore.RegisteredClaims
}

func TestMiddlewareBuilder_Build(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name            string
		reqBuilder      func(t *testing.T) *http.Request
		isUseMiddleware bool
		wantCode        int
	}
	tests := []testCase[Claims, *Claims]{
		{
			// 未使用中间件
			name: "no_middleware_used",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			isUseMiddleware: false,
			wantCode:        http.StatusOK,
		},
		{
			// 使用中间件但没有 token
			name: "use_middleware_and_no_token",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			isUseMiddleware: true,
			wantCode:        http.StatusUnauthorized,
		},
		{
			// 使用中间件且通过认证
			name: "use_middleware_and_pass_authentication",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add(authorizationHeader, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.B9sIBtCtX5kp8pk0fjpcy-8HVa991qU5L5nles7Nblw")
				return req
			},
			isUseMiddleware: true,
			wantCode:        http.StatusOK,
		},
		{
			// 使用中间件但 token 错误未通过认证
			name: "use_middleware_and_fail_authentication",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add(authorizationHeader, "Bearer bad_token")
				return req
			},
			isUseMiddleware: true,
			wantCode:        http.StatusUnauthorized,
		},
		{
			// 使用中间件但 token 格式错误未通过认证
			name: "use_middleware_and_fail_authentication",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add(authorizationHeader, "BearereyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.B9sIBtCtX5kp8pk0fjpcy-8HVa991qU5L5nles7Nblw")
				return req
			},
			isUseMiddleware: true,
			wantCode:        http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims, *Claims](tokenManager)
			server := gin.Default()
			if tt.isUseMiddleware {
				server.Use(m.Build())
			}
			m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func TestMiddlewareBuilder_IgnorePathFunc(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name    string
		fn      func(*gin.Context) bool
		isUseFn bool
		req     func(t *testing.T) *http.Request
		want    bool
	}
	tests := []testCase[Claims, *Claims]{
		{
			name: "normal",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			want: false,
		},
		{
			name: "another",
			fn: func(c *gin.Context) bool {
				return c.Request.URL.Path == "/ok"
			},
			isUseFn: true,
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/ok", nil)
				require.NoError(t, err)
				return req
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims, *Claims](tokenManager)
			if tt.isUseFn {
				m.IgnorePathFunc(tt.fn)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = tt.req(t)
			got := m.ignorePath(c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMiddlewareBuilder_SetExtractTokenFunc(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name    string
		fn      func(*gin.Context) string
		isUseFn bool
		req     func(t *testing.T) *http.Request
		want    string
	}
	tests := []testCase[Claims, *Claims]{
		{
			name: "normal",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add(authorizationHeader, "Bearer token")
				return req
			},
			want: "token",
		},
		{
			name: "another",
			fn: func(c *gin.Context) string {
				return "fixed token"
			},
			isUseFn: true,
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				req.Header.Add(authorizationHeader, "Bearer token")
				return req
			},
			want: "fixed token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims, *Claims](tokenManager)
			if tt.isUseFn {
				m.SetExtractTokenFunc(tt.fn)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = tt.req(t)
			got := m.extractToken(c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMiddlewareBuilder_SetClaimsFunc(t *testing.T) {
	type clmKey struct{}
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name    string
		fn      func(*gin.Context, T)
		isUseFn bool
		req     func(t *testing.T) *http.Request
		clm     T
		after   func(t *testing.T, c *gin.Context)
	}
	tests := []testCase[Claims, *Claims]{
		{
			name: "normal",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			clm: Claims{
				Uid: 123,
			},
			after: func(t *testing.T, c *gin.Context) {
				tmp, _ := c.Get(claimsKey)
				clm, ok := tmp.(Claims)
				require.Equal(t, true, ok)
				assert.Equal(t, int64(123), clm.Uid)
			},
		},
		{
			name: "another",
			fn: func(c *gin.Context, claims Claims) {
				ctx := context.WithValue(c.Request.Context(), clmKey{}, claims)
				c.Request = c.Request.WithContext(ctx)
			},
			isUseFn: true,
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				require.NoError(t, err)
				return req
			},
			clm: Claims{
				Uid: 456,
			},
			after: func(t *testing.T, c *gin.Context) {
				tmp := c.Request.Context().Value(clmKey{})
				clm, ok := tmp.(Claims)
				require.Equal(t, true, ok)
				assert.Equal(t, int64(456), clm.Uid)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims, *Claims](tokenManager)
			if tt.isUseFn {
				m.SetClaimsFunc(tt.fn)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = tt.req(t)
			m.setClaims(c, tt.clm)
			tt.after(t, c)
		})
	}
}

func TestMiddlewareBuilder_IgnoreFullPath(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name       string
		fullPaths  []string
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
	}
	tests := []testCase[Claims, *Claims]{
		{
			name:      "normal",
			fullPaths: []string{"/login", "/user/:id"},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "normal_by_full_path",
			fullPaths: []string{"/login", "/user/:id"},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/user/1", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "incorrect_full_path",
			fullPaths: []string{"/login", "/user/:id"},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/user/1/detial", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims, *Claims](tokenManager)
			server := gin.Default()
			server.Use(m.IgnoreFullPath(tt.fullPaths...).Build())
			m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

var (
	tokenManager = jwtcore.NewTokenManagerServer[Claims, *Claims](
		10*time.Minute, "sign key",
		jwtcore.WithTimeFunc[Claims, *Claims](func() time.Time {
			return time.UnixMilli(1695571200000)
		}))
)

func (m *MiddlewareBuilder[T, PT]) registerRoutes(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/user/:id", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/login", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
