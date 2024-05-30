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
	"github.com/stretchr/testify/require"
)

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

func TestBuilder_Build(t *testing.T) {
	const limitURL = "/limit"
	mockLimiter := &mockLimiter{}
	svc := NewBuilder(mockLimiter)
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
				require.NoError(t, err)
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
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			// 超时
			name:       "ctx_timeout_error",
			limited:    true,
			limiterErr: context.DeadlineExceeded,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusGatewayTimeout,
		},
		{
			// 限流器增加时错误
			name:       "limiter_add_error",
			limited:    false,
			limiterErr: errors.New("模拟限流器错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLimiter.limited = tt.limited
			mockLimiter.limitErr = tt.limiterErr
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

func TestBuilder_BuildBlock(t *testing.T) {
	const limitURL = "/limit"
	mockLimiter := &mockLimiter{}
	svc := NewBuilder(mockLimiter)
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
				require.NoError(t, err)
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
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			// 超时
			name:       "ctx_timeout_error",
			limited:    true,
			limiterErr: context.DeadlineExceeded,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusGatewayTimeout,
		},
		{
			// 限流器增加时错误
			name:       "limiter_add_error",
			limited:    false,
			limiterErr: errors.New("模拟限流器错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLimiter.blockLimited = tt.limited
			mockLimiter.blockLimitErr = tt.limiterErr
			server := gin.Default()
			server.Use(svc.BuildBlock())
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
	limited       bool
	limitErr      error
	blockLimited  bool
	blockLimitErr error
}

func (l *mockLimiter) Put() {}

func (l *mockLimiter) Close() {}

func (l *mockLimiter) Limit(_ context.Context, _ string) (bool, error) {
	return l.limited, l.limitErr
}

func (l *mockLimiter) BlockLimit(_ context.Context, _ string) (bool, error) {
	return l.blockLimited, l.blockLimitErr
}
