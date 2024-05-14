package limit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_SetKeyGenFunc(t *testing.T) {
	tests := []struct {
		name       string
		reqBuilder func(t *testing.T) *http.Request
		fn         func(*gin.Context) string
		want       string
	}{
		{
			// 设置key成功
			name: "set_key_success",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			fn: func(ctx *gin.Context) string {
				return "test"
			},
			want: "test",
		},
		{
			// 默认key
			name: "default_key",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: "ip-limiter:127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(nil)
			if tt.fn != nil {
				b.SetKeyGenFunc(tt.fn)
			}

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tt.reqBuilder(t)
			ctx.Request = req

			assert.Equal(t, tt.want, b.genKeyFn(ctx))
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	const limitURL = "/limit"
	testLimiter := &testLimiter{}
	svc := NewBuilder(testLimiter)
	tests := []struct {
		name       string
		limited    bool
		limiterErr error
		reqBuilder func(t *testing.T) *http.Request
		// 预期响应
		wantCode int
	}{
		{
			// 不限流
			name:       "no_limit",
			limited:    false,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			// 限流
			name:       "limited",
			limited:    true,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			// 系统错误
			name:       "system_error",
			limited:    false,
			limiterErr: errors.New("模拟系统错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLimiter.limited = tt.limited
			testLimiter.err = tt.limiterErr
			server := gin.Default()
			server.Use(svc.Build())
			svc.RegisterRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func TestBuilder_limit(t *testing.T) {
	testLimiter := &testLimiter{}
	tests := []struct {
		name       string
		limited    bool
		limiterErr error
		reqBuilder func(t *testing.T) *http.Request
		// 预期响应
		want    bool
		wantErr error
	}{
		{
			name: "不限流",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: false,
		},
		{
			name:       "限流",
			limited:    true,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: true,
		},
		{
			name:       "限流代码出错",
			limited:    false,
			limiterErr: errors.New("模拟系统错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want:    false,
			wantErr: errors.New("模拟系统错误"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLimiter.limited = tt.limited
			testLimiter.err = tt.limiterErr
			b := NewBuilder(testLimiter)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tt.reqBuilder(t)
			ctx.Request = req

			got, err := b.limit(ctx)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func (b *Builder) RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

func TestBuilder_SetLogger(t *testing.T) {
	var l *slog.Logger
	tests := []struct {
		name string
		fn   *slog.Logger
		want *slog.Logger
	}{
		{
			name: "normal",
			fn:   l,
			want: l,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(&testLimiter{})
			b.SetLogger(tt.fn).Build()
			assert.Equal(t, tt.want, b.logger)
		})
	}
}

type testLimiter struct {
	limited bool
	err     error
}

func (t *testLimiter) Limit(_ context.Context, _ string) (bool, error) {
	return t.limited, t.err
}
