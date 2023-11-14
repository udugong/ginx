package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	ujwt "github.com/udugong/ginx/jwt"
)

type jwtData struct {
	Uid int64 `json:"uid"`
}

func main() {
	r := gin.Default()

	accessKey := "access key"
	m := ujwt.NewManagement[jwtData](ujwt.NewOptions(10*time.Minute, accessKey))

	// 登录认证中间件
	// gloAuthMiddleware := m.MiddlewareBuilder().IgnorePath("/login", "/signup").Build()
	authMiddleware := m.MiddlewareBuilder().Build()

	// 全局拦截
	// r.Use(gloAuthMiddleware)

	// 单独拦截
	r.GET("/profile", authMiddleware, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"userName": "foo"})
	})

	// 登录设置资源令牌
	r.POST("/login", func(ctx *gin.Context) {
		// ...
		// 如果校验成功
		token, err := m.GenerateAccessToken(jwtData{Uid: 1})
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Header("x-access-token", token)
		ctx.Status(http.StatusNoContent)
	})

	// 使用刷新令牌相关内容需要设置 refreshJWTOptions
	refreshKey := "refresh key"
	m = ujwt.NewManagement[jwtData](
		ujwt.NewOptions(10*time.Minute, accessKey),
		ujwt.WithRefreshJWTOptions[jwtData](
			ujwt.NewOptions(7*24*time.Hour, refreshKey)),
		// 开启轮换刷新令牌(Refresh 的时候会生成一个新的 refresh token)
		ujwt.WithRotateRefreshToken[jwtData](true),
	)

	// 登录
	r.POST("/login-v1", func(ctx *gin.Context) {
		// ...
		// 如果校验成功
		accessToken, err := m.GenerateAccessToken(jwtData{Uid: 1})
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		refreshToken, err := m.GenerateRefreshToken(jwtData{Uid: 1})
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Header("x-access-token", accessToken)
		ctx.Header("x-refresh-token", refreshToken)
		ctx.Status(http.StatusNoContent)
	})

	// 刷新令牌的函数
	r.POST("/refresh-token", m.Refresh)

	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
