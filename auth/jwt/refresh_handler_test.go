package jwt

import (
	"context"
	"encoding/json"
	"errors"
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

func TestRefreshManager_Handler(t *testing.T) {
	accessKey := "access key"
	refreshKey := "refresh key"
	defaultExpire := 10 * time.Minute
	nowTime := time.UnixMilli(1695571200000)
	accessTM := jwtcore.NewTokenManager[Claims](
		accessKey, defaultExpire,
		jwtcore.WithTimeFunc[Claims](func() time.Time {
			return nowTime
		}),
		jwtcore.WithAddParserOption[Claims](jwt.WithTimeFunc(
			func() time.Time { return nowTime },
		)),
	)
	refreshTM := jwtcore.NewTokenManager[Claims](
		refreshKey, 24*time.Hour,
		jwtcore.WithTimeFunc[Claims](func() time.Time {
			return nowTime
		}),
		jwtcore.WithAddParserOption[Claims](jwt.WithTimeFunc(
			func() time.Time { return nowTime },
		)),
	)

	type testCase[T jwt.Claims] struct {
		name       string
		h          *RefreshManager[T]
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		after      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}
	tests := []testCase[Claims]{
		{
			// 更新资源令牌并轮换刷新令牌
			name: "refresh_access_token_and_rotate_refresh_token",
			h: NewRefreshManager[Claims](accessTM, refreshTM,
				WithRotateRefreshToken[Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusNoContent,
			after: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				accessToken := recorder.Header().Get("x-access-token")
				refreshToken := recorder.Header().Get("x-refresh-token")
				assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.Azhc3P_Iks_DRWRZUrZwpKWLiZ9LY7fI0BqhLzOsEgI",
					accessToken)
				assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzYwMCwiaWF0IjoxNjk1NTcxMjAwfQ.USVVhRntQtzwblLWSrImY2PpxRkYpyxEycMeVc4UVhs",
					refreshToken)
			},
		},
		{
			// 更新资源令牌但轮换刷新令牌生成失败
			name: "refresh_access_token_but_gen_rotate_refresh_token_failed",
			h: NewRefreshManager[Claims](accessTM, &testTokenManager{
				generateErr:  errors.New("模拟生成 refresh token 失败"),
				verifyClaims: Claims{Uid: 1},
				verifyErr:    nil,
			}, WithRotateRefreshToken[Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
			after:    func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
		{
			// 仅更新资源令牌
			name: "refresh_access_token",
			h:    NewRefreshManager[Claims](accessTM, refreshTM),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusNoContent,
			after: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				accessToken := recorder.Header().Get("x-access-token")
				assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.Azhc3P_Iks_DRWRZUrZwpKWLiZ9LY7fI0BqhLzOsEgI", accessToken)
			},
		},
		{
			// 生成资源令牌失败
			name: "gen_access_token_failed",
			h: NewRefreshManager[Claims](&testTokenManager{
				generateErr:  errors.New("模拟生成 access token 失败"),
				verifyClaims: Claims{Uid: 1},
			}, refreshTM),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
			after:    func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
		{
			// 获取 claims 失败
			name: "failed_to_obtain_claims",
			h: NewRefreshManager[Claims](accessTM, refreshTM,
				WithRefreshAuthHandler[Claims](func(c *gin.Context) {
					return
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusInternalServerError,
			after:    func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
		{
			// 认证失败直接中断执行
			name: "unauthorized",
			h: NewRefreshManager[Claims](accessTM, refreshTM,
				WithRotateRefreshToken[Claims](true)),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer bad_token")
				return req
			},
			wantCode: http.StatusUnauthorized,
			after:    func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
		{
			name: "change_option",
			h: NewRefreshManager[Claims](accessTM, refreshTM,
				WithRefreshAuthHandler[Claims](
					NewMiddlewareBuilder[Claims](refreshTM).SetClaimsFunc(
						func(c *gin.Context, claims Claims) {
							ctx := context.WithValue(c.Request.Context(), "claims", claims)
							c.Request = c.Request.WithContext(ctx)
						},
					).Build(),
				),
				WithGetClaims[Claims](func(c *gin.Context) (Claims, bool) {
					v := c.Request.Context().Value("claims")
					clm, ok := v.(Claims)
					return clm, ok
				}),
				WithAccessTokenSetter[Claims](func(c *gin.Context, token string) {
					ctx := context.WithValue(c.Request.Context(), "access-token", token)
					c.Request = c.Request.WithContext(ctx)
				}),
				WithRotateRefreshToken[Claims](true),
				WithRefreshTokenSetter[Claims](func(c *gin.Context, token string) {
					ctx := context.WithValue(c.Request.Context(), "refresh-token", token)
					c.Request = c.Request.WithContext(ctx)
				}),
				WithResponseSetter[Claims](func(c *gin.Context) {
					v1 := c.Request.Context().Value("access-token")
					accessToken, ok := v1.(string)
					if !ok {
						c.Status(http.StatusInternalServerError)
						return
					}

					v2 := c.Request.Context().Value("refresh-token")
					refreshToken, ok := v2.(string)
					if !ok {
						c.Status(http.StatusInternalServerError)
						return
					}

					c.JSON(http.StatusOK, gin.H{
						"access_token":  accessToken,
						"refresh_token": refreshToken,
					})
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				require.NoError(t, err)
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzQwMCwiaWF0IjoxNjk1NTcxMDAwfQ.gew4g8GdYdl3COOeHh5AmnnSAA3tgJ8WWkV3GI6cILQ")
				return req
			},
			wantCode: http.StatusOK,
			after: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				respMap := map[string]string{
					"access_token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTU3MTgwMCwiaWF0IjoxNjk1NTcxMjAwfQ.Azhc3P_Iks_DRWRZUrZwpKWLiZ9LY7fI0BqhLzOsEgI",
					"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTY5NTY1NzYwMCwiaWF0IjoxNjk1NTcxMjAwfQ.USVVhRntQtzwblLWSrImY2PpxRkYpyxEycMeVc4UVhs",
				}
				b, err := json.Marshal(respMap)
				assert.NoError(t, err)
				assert.Equal(t, string(b), recorder.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			server.GET("/refresh", tt.h.Handler)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.wantCode != recorder.Code {
				return
			}
			tt.after(t, recorder)
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

func (m *testTokenManager) VerifyToken(_ string) (Claims, error) {
	return m.verifyClaims, m.verifyErr
}
