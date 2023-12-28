package jwt

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/udugong/ginx/jwt/jwtcore"
)

func TestRefreshHandlerBuilder_Build(t *testing.T) {
	defaultExpire := 10 * time.Minute
	accessKey := "access key"
	refreshKey := "refresh key"
	nowTime := time.UnixMilli(1695571200000)
	accessTM := jwtcore.NewTokenManagerServer[Claims, *Claims](
		defaultExpire, accessKey,
		jwtcore.WithTimeFunc[Claims, *Claims](func() time.Time {
			return nowTime
		}))
	refreshTM := jwtcore.NewTokenManagerServer[Claims, *Claims](
		24*time.Hour, refreshKey,
		jwtcore.WithTimeFunc[Claims, *Claims](func() time.Time {
			return nowTime
		}))

	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name             string
		h                *RefreshHandlerBuilder[T, PT]
		reqBuilder       func(t *testing.T) *http.Request
		wantCode         int
		wantAccessToken  string
		wantRefreshToken string
	}
	tests := []testCase[Claims, *Claims]{
		{
			// 更新资源令牌并轮换刷新令牌
			name: "refresh_access_token_and_rotate_refresh_token",
			h: NewRefreshHandlerBuilder[Claims, *Claims](accessTM, refreshTM,
				WithRotateRefreshToken[Claims, *Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode:         http.StatusNoContent,
			wantAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.Azhc3P_Iks_DRWRZUrZwpKWLiZ9LY7fI0BqhLzOsEgI",
			wantRefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzYwMCwiaWF0IjoxNjk1NTcxMjAwfQ.USVVhRntQtzwblLWSrImY2PpxRkYpyxEycMeVc4UVhs",
		},
		{
			// 更新资源令牌但轮换刷新令牌生成失败
			name: "refresh_access_token_but_gen_rotate_refresh_token_failed",
			h: NewRefreshHandlerBuilder[Claims, *Claims](accessTM, &testTokenManager{
				generateErr:  errors.New("模拟生成 refresh token 失败"),
				verifyClaims: Claims{Uid: 1},
				verifyErr:    nil,
			}, WithRotateRefreshToken[Claims, *Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 仅更新资源令牌
			name: "refresh_access_token",
			h:    NewRefreshHandlerBuilder[Claims, *Claims](accessTM, refreshTM),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode:        http.StatusNoContent,
			wantAccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.Azhc3P_Iks_DRWRZUrZwpKWLiZ9LY7fI0BqhLzOsEgI",
		},
		{
			// 生成资源令牌失败
			name: "gen_access_token_failed",
			h: NewRefreshHandlerBuilder[Claims, *Claims](&testTokenManager{
				generateErr:  errors.New("模拟生成 access token 失败"),
				verifyClaims: Claims{Uid: 1},
			}, refreshTM),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 获取 claims 失败
			name: "failed_to_obtain_claims",
			h: &RefreshHandlerBuilder[Claims, *Claims]{
				refreshAuthHandler: func(c *gin.Context) { return },
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 认证失败直接中断执行
			name: "unauthorized",
			h: NewRefreshHandlerBuilder[Claims, *Claims](accessTM, refreshTM,
				WithRotateRefreshToken[Claims, *Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer bad_token")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			server.GET("/refresh", tt.h.Build)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.wantCode != http.StatusNoContent {
				return
			}
			assert.Equal(t, tt.wantAccessToken,
				recorder.Header().Get("x-access-token"))
			assert.Equal(t, tt.wantRefreshToken,
				recorder.Header().Get("x-refresh-token"))
		})
	}
}

func TestWithExposeAccessHeader(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want string
	}
	tests := []testCase[Claims, *Claims]{
		{
			name: "normal",
			fn:   withNop[Claims, *Claims],
			want: "x-access-token",
		},
		{
			name: "set_another_access_token",
			fn: func() Option[Claims, *Claims] {
				return WithExposeAccessHeader[Claims, *Claims]("access-token")
			},
			want: "access-token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRefreshHandlerBuilder[Claims, *Claims](
				nil, nil, tt.fn()).exposeAccessHeader
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeRefreshHeader(t *testing.T) {
	type testCase[T jwt.Claims, PT jwtcore.Claims[T]] struct {
		name string
		fn   func() Option[T, PT]
		want string
	}
	tests := []testCase[Claims, *Claims]{
		{
			name: "normal",
			fn:   withNop[Claims, *Claims],
			want: "x-refresh-token",
		},
		{
			name: "set_another_refresh-token",
			fn: func() Option[Claims, *Claims] {
				return WithExposeRefreshHeader[Claims, *Claims]("refresh-token")
			},
			want: "refresh-token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRefreshHandlerBuilder[Claims, *Claims](
				nil, nil, tt.fn()).exposeRefreshHeader
			assert.Equal(t, tt.want, got)
		})
	}
}

type testTokenManager struct {
	generateToken string
	generateErr   error
	verifyClaims  Claims
	verifyErr     error
}

func (m *testTokenManager) GenerateToken(_ Claims) (string, error) {
	return m.generateToken, m.generateErr
}

func (m *testTokenManager) VerifyToken(_ string, _ ...jwt.ParserOption) (Claims, error) {
	return m.verifyClaims, m.verifyErr
}

func withNop[T jwt.Claims, PT jwtcore.Claims[T]]() Option[T, PT] {
	return optionFunc[T, PT](func(r *RefreshHandlerBuilder[T, PT]) {})
}
