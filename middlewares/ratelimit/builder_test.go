package ratelimit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/udugong/ginx/internal/ratelimit"
	limitmocks "github.com/udugong/ginx/internal/ratelimit/mocks"
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
	tests := []struct {
		name string

		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request

		// 预期响应
		wantCode int
	}{
		{
			// 不限流
			name: "no_limit",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
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
			name: "limited",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
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
			name: "system_error",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewBuilder(tt.mock(ctrl))

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
	tests := []struct {
		name string

		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request

		// 预期响应
		want    bool
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
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
			name: "限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
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
			name: "限流代码出错",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limiter := tt.mock(ctrl)
			b := NewBuilder(limiter)

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

func TestBuilder_SetLogFunc(t *testing.T) {
	var l func(msg any, args ...any)
	tests := []struct {
		name string
		fn   func(msg any, args ...any)
		want *func(msg any, args ...any)
	}{
		{
			name: "normal",
			fn:   l,
			want: &l,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(NewRedisSlidingWindowLimiter(nil, time.Second, 100))
			b.SetLogFunc(tt.fn).Build()
			assert.Equal(t, tt.want, &b.logFn)
		})
	}
}
