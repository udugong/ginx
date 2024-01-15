package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	ujwt "github.com/udugong/ginx/jwt"
	"github.com/udugong/ginx/jwt/jwtcore"
)

type Claims struct {
	Uid int64 `json:"uid"`
	// 嵌入
	jwtcore.RegisteredClaims
}

func main() {
	r := gin.Default()

	accessKey := "access key"
	// 创建资源令牌管理服务
	accessTM := jwtcore.NewTokenManagerServer[Claims, *Claims](accessKey, 10*time.Minute)
	// 创建认证中间件构建器
	builder := ujwt.NewMiddlewareBuilder[Claims, *Claims](accessTM)

	// // 构建全局登录认证中间件
	// gloAuthMiddleware := builder.
	// 	// 忽略 "/login", "/signup" 这两个 full path 的认证。
	// 	IgnoreFullPath("/login", "/signup").Build()
	// // 全局拦截
	// r.Use(gloAuthMiddleware)

	// 单独拦截
	r.GET("/profile", builder.Build(), func(c *gin.Context) {
		// 获取 claims
		v, _ := c.Get("claims") // 默认存放在 key = "claims" 的 gin.Context 中.
		clm, ok := v.(Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, "不应该命中该分支")
			return
		}
		fmt.Println(clm.Uid) // 获取 Uid
		c.Status(http.StatusOK)
	})

	// 登录设置资源令牌
	r.POST("/login", func(c *gin.Context) {
		// ...
		// 如果校验成功
		token, err := accessTM.GenerateToken(Claims{Uid: 1})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("x-access-token", token)
		c.Status(http.StatusNoContent)
	})

	// 使用刷新令牌处理函数
	refreshKey := "refresh key"
	// 创建刷新令牌的管理服务
	refreshTM := jwtcore.NewTokenManagerServer[Claims, *Claims](refreshKey, 24*time.Hour)
	// 创建刷新令牌函数的构建器
	refreshManager := ujwt.NewRefreshManager[Claims, *Claims](accessTM, refreshTM)

	// 登录
	r.POST("/login-v1", func(c *gin.Context) {
		// ...
		// 如果校验成功
		accessToken, err := accessTM.GenerateToken(Claims{Uid: 1})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		refreshToken, err := refreshTM.GenerateToken(Claims{Uid: 1})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("x-access-token", accessToken)
		c.Header("x-refresh-token", refreshToken)
		c.Status(http.StatusNoContent)
	})

	// 刷新令牌的函数
	// 内部已经开启了 refresh 令牌的认证,因此可以直接使用 refreshManager.Handler 与 relativePath 绑定
	r.POST("/refresh-token", refreshManager.Handler)

	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
