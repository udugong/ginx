package activelimit

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
			want: "all_req_active_limiter",
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
			b := NewBuilder(&mockLimiter{})
			b.SetLogger(tt.fn).Build()
			assert.Equal(t, tt.want, b.logger)
		})
	}
}

func TestBuilder_SetKeyGenFuncByIP(t *testing.T) {
	tests := []struct {
		name       string
		reqBuilder func(t *testing.T) *http.Request
		useIP      bool
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
			useIP: true,
			want:  "ip_active_limiter:127.0.0.1",
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
			want: "all_req_active_limiter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(nil)
			if tt.useIP {
				b.SetKeyGenFuncByIP()
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
	mockLimiter := &mockLimiter{}
	svc := NewBuilder(mockLimiter)
	tests := []struct {
		name       string
		limited    bool
		limiterErr error
		decrErr    error
		reqBuilder func(t *testing.T) *http.Request
		// 预期响应
		wantCode int
	}{
		{
			// 不限流
			name:       "no_limit",
			limited:    false,
			limiterErr: nil,
			decrErr:    nil,
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
			decrErr:    nil,
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
			// 限流器增加时错误
			name:       "limiter_add_error",
			limited:    false,
			limiterErr: errors.New("模拟限流器错误"),
			decrErr:    nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 限流器时扣减错误
			name:       "limiter_decr_error",
			limited:    false,
			limiterErr: nil,
			decrErr:    errors.New("模拟限流器错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLimiter.limited = tt.limited
			mockLimiter.limitErr = tt.limiterErr
			mockLimiter.decrErr = tt.decrErr
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

func (b *Builder) RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

type mockLimiter struct {
	limited  bool
	limitErr error
	decrErr  error
}

func (l *mockLimiter) Limit(_ context.Context, _ string) (bool, error) {
	return l.limited, l.limitErr
}

func (l *mockLimiter) Decr(_ context.Context, _ string) error {
	return l.decrErr
}
