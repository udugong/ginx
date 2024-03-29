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
	"github.com/udugong/token/jwtcore"
)

type Claims struct {
	Uid int64 `json:"uid,omitempty"`
	jwtcore.RegisteredClaims
}

func TestMiddlewareBuilder_Build(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name            string
		reqBuilder      func(t *testing.T) *http.Request
		isUseMiddleware bool
		wantCode        int
	}
	tests := []testCase[Claims]{
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
			m := NewMiddlewareBuilder[Claims](tokenManager)
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
	type testCase[T jwt.Claims] struct {
		name    string
		fn      func(*gin.Context) bool
		isUseFn bool
		req     func(t *testing.T) *http.Request
		want    bool
	}
	tests := []testCase[Claims]{
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
			m := NewMiddlewareBuilder[Claims](tokenManager)
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
	type testCase[T jwt.Claims] struct {
		name    string
		fn      func(*gin.Context) string
		isUseFn bool
		req     func(t *testing.T) *http.Request
		want    string
	}
	tests := []testCase[Claims]{
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
			m := NewMiddlewareBuilder[Claims](tokenManager)
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
	clmKey := "claims"
	type testCase[T jwt.Claims] struct {
		name    string
		fn      func(*gin.Context, T)
		isUseFn bool
		req     func(t *testing.T) *http.Request
		clm     T
		after   func(t *testing.T, c *gin.Context)
	}
	tests := []testCase[Claims]{
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
				clm, ok := ClaimsFromContext[Claims](c.Request.Context())
				assert.True(t, ok)
				assert.Equal(t, int64(123), clm.Uid)
			},
		},
		{
			name: "another",
			fn: func(c *gin.Context, claims Claims) {
				c.Set(clmKey, claims)
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
				tmp, _ := c.Get(clmKey)
				clm, ok := tmp.(Claims)
				require.Equal(t, true, ok)
				assert.Equal(t, int64(456), clm.Uid)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims](tokenManager)
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
	type testCase[T jwt.Claims] struct {
		name       string
		fullPaths  []string
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
	}
	tests := []testCase[Claims]{
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
				req, err := http.NewRequest(http.MethodGet, "/user/1/detail", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddlewareBuilder[Claims](tokenManager)
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

func TestContextWithClaims(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name   string
		ctx    context.Context
		claims T
		want   context.Context
	}
	tests := []testCase[Claims]{
		{
			name:   "normal",
			ctx:    context.Background(),
			claims: Claims{Uid: 1},
			want:   context.WithValue(context.Background(), claimsKey{}, Claims{Uid: 1}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ContextWithClaims(tt.ctx, tt.claims))
		})
	}
}

func TestClaimsFromContext(t *testing.T) {
	type testCase[T jwt.Claims] struct {
		name  string
		ctx   context.Context
		want  T
		want1 bool
	}
	tests := []testCase[Claims]{
		{
			name:  "normal",
			ctx:   context.WithValue(context.Background(), claimsKey{}, Claims{Uid: 1}),
			want:  Claims{Uid: 1},
			want1: true,
		},
		{
			name:  "claims_not_found",
			ctx:   context.Background(),
			want:  Claims{},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ClaimsFromContext[Claims](tt.ctx)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

var (
	tokenManager = jwtcore.NewTokenManager[Claims, *Claims](
		"sign key", 10*time.Minute,
		jwtcore.WithAddParserOption[Claims](jwt.WithTimeFunc(func() time.Time {
			return time.UnixMilli(1695571200000)
		})))
)

func (m *MiddlewareBuilder[T]) registerRoutes(server *gin.Engine) {
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
