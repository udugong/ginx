# Ginx
gin的插件



Go Versions
==================

`>=1.20`



# Usage

下载安装：`go get github.com/udugong/ginx`

  * [jwt 的使用](#jwt-package)
  * [ratelimit 限流](#ratelimit-限流)



# `jwt` package

该`jwt`包提供了一些有用的方法，使您可以在使用 gin 时快速完成认证功能。

- 利用泛型可以自定义 claims 内容
- 生成/校验 token（在 `jwtcore` 包中）
- 登录认证中间件
- 刷新 token 的 gin.HandlerFunc

```go
package main

import (
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

```

注意:
- 关于请求头CORS的问题可以查看[cors](https://github.com/gin-contrib/cors)中间件解决。
- 用户认证中间件默认是根据`Authorization`请求头内容来进行校验。需要在`cors.Config`中配置`AllowHeaders`。
- `RefreshManager`中的 Handler 默认把令牌都放到`x-access-token`和`x-refresh-token`请求头中。需要在`cors.Config`中配置`ExposeHeaders`。



# `ratelimit` 限流

ratelimit 是基于 `redis` 实现的滑动窗口限流器。它为 gin 提供了限流中间件，使您快速完成针对 IP 的限流，您也可以设置不同的 key 来实现不同的限流。

```go
package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/udugong/ginx/middlewares/ratelimit"
)

func main() {
	rdb := InitRedis()
	// 创建一个基于 redis, 1000/s 的限流器
	limiter := ratelimit.NewRedisSlidingWindowLimiter(rdb, time.Second, 1000)
	builder := ratelimit.NewBuilder(limiter)

	r := gin.Default()
	// 控制所有的请求的速率
	r.Use(builder.SetKeyGenFunc(func(*gin.Context) string {
		return "all-request" // 设置 redis 的 key
	}).Build())

	// 默认是根据 IP 限流
	// 每个 IP 每秒 最多访问 1000次
	r.Use(builder.Build())
}

func InitRedis() redis.Cmdable {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	return rdb
}

```

