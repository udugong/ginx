# ginx

gin的插件



go versions
==================

`>=1.21`

# usage

下载安装：`go get github.com/udugong/ginx@latest`

* [auth 认证](#auth-package)
* [limit 限流](#limit-package)



# `auth` package

该`auth`包提供了一些有用的方法，使您可以在使用 gin 时快速完成用户认证功能。

- [jwt 认证](#jwt-认证)

## jwt 认证

该`jwt`包提供了 jwt 认证功能。该功能借助了 [token/jwtcore](https://github.com/udugong/token) 包的 jwt 生成/校验功能实现。

- 利用泛型可以自定义 claims 内容
- 登录认证中间件
- 刷新 token 的 gin.HandlerFunc

#### 使用方法

注意:

- 关于请求头CORS的问题可以查看[cors](https://github.com/gin-contrib/cors)中间件解决。
- 使用该中间件以及刷新函数，默认会从 `Authorization` 请求头中获取 `Bearer xxxx` 来进行校验。需要在  `cors.Config.AllowHeaders` 中添加 `authorization` 。
- `RefreshManager`中的 Handler 默认把令牌都放到 `X-Access-Token` 和 `X-Refresh-Token` 请求头中。需要在 `cors.Config.ExposeHeaders` 中添加这两个请求头，前端才能正常获取。

1. 定义 Claims 结构体

   需要嵌入 [jwtcore.RegisteredClaims](https://github.com/udugong/token/blob/main/jwtcore/registered_claims.go#L14) 或者实现 [jwtcore.Claims[T jwt.Claims]](https://github.com/udugong/token/blob/main/jwtcore/claims.go#L11) 接口，其余成员可以自行定义。
   
   ```go
   import "github.com/udugong/token/jwtcore"

   type Claims struct {
   	Uid int64 `json:"uid"` // 用户ID
   	// Nickname string `json:"nickname"` // 昵称
   	// 嵌入
   	jwtcore.RegisteredClaims
   }
   ```
   
2. 创建令牌管理器

   通过 jwtcore 创建一个令牌管理器。详细参考 [token/jwtcore](https://github.com/udugong/token) 。

   ```go
   // 创建资源令牌管理器
   accessKey := "access key"
   accessTM := jwtcore.NewTokenManager[Claims](accessKey, 10*time.Minute)
   ```
   
3. 创建认证中间件

   需要传入一个令牌管理器，使用 `Build()` 创建 `gin` 的中间件。您可以灵活的构建。

   ```go
   import (
   	"github.com/gin-gonic/gin"
   	ujwt "github.com/udugong/ginx/auth/jwt"
   	"github.com/udugong/token/jwtcore"
   )
   
   // 创建认证中间件构建器
   builder := ujwt.NewMiddlewareBuilder[Claims](accessTM)
   gin.Default().Use(builder.Build()) // 使用中间件
   ```

   您可以使用`jwt`包提供的方法灵活构建。

   ```go
   // 忽略 "/login", "/signup" 这两个 full path 的认证。
   builder.IgnoreFullPath("/login", "/signup").Build()
   ```

4. 使用刷新令牌的 gin.HandlerFunc

   需要创建一个刷新令牌的管理器。创建刷新令牌函数的构建器时需要注意：插入 Claims 的具体类型。

   ```go
   // 使用刷新令牌处理函数
   refreshKey := "refresh key"
   // 创建刷新令牌的管理服务
   refreshTM := jwtcore.NewTokenManager[Claims](refreshKey, 24*time.Hour)
   // 创建刷新令牌函数的构建器
   refreshManager := ujwt.NewRefreshManager[Claims](accessTM, refreshTM)
   
   // 内部已经开启了 refresh 令牌的认证,因此可以直接使用 refreshManager.Handler 与 relativePath 绑定
   gin.Default().POST("/refresh-token", refreshManager.Handler)
   ```
   
   您可以使用`jwt`包提供的方法灵活构建。以下提供一些示例：
   
   - 轮换 refresh token
   
     每次调用刷新令牌函数会重新生成新的 `refresh token` 。默认不开启轮换刷新令牌。如需要轮换 refresh token 则需要把 `WithRotateRefreshToken(true)` 传入以构建刷新令牌管理器。
   
     ```go
     ujwt.NewRefreshManager[Claims](accessTM, refreshTM, ujwt.WithRotateRefreshToken[Claims](true))
     ```
   
   - 修改响应
   
     允许自定义刷新令牌函数的响应。传入 `WithResponseSetter` 可以更改响应中。最佳的方式应该是配合 `WithAccessTokenSetter` 和 `WithRefreshTokenSetter` 一同与 `WithResponseSetter` 操作。
   
     ```go
     ujwt.NewRefreshManager[Claims](accessTM, refreshTM,
     	ujwt.WithRotateRefreshToken[Claims](true), // 允许轮换 refresh token
     	// 修改响应把 token 放入 Body 中
     	ujwt.WithResponseSetter[Claims](func(c *gin.Context) {
     		accessToken := c.Writer.Header().Get("x-access-token")
     		refreshToken := c.Writer.Header().Get("x-refresh-token")
     		c.Writer.Header().Del("x-access-token")
     		c.Writer.Header().Del("x-refresh-token")
     		c.JSON(http.StatusOK, gin.H{
     			"access-token":  accessToken,
     			"refresh-token": refreshToken,
     		})
     	}),
     )
     ```
   
5. 获取 Claims

   Claims 默认存放在 context.Context 中。对外提供了 `ClaimsFromContext` 方法获取 Claims。如果不存在 Claims 或者类型错误则会返回 false。

   ```go
   ujwt.ClaimsFromContext[Claims](c.Request.Context()) // context.Context 存放于 *gin.Context.Request.Context() 中
   ```

#### 示例

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ujwt "github.com/udugong/ginx/auth/jwt"
	"github.com/udugong/token/jwtcore"
)

type Claims struct {
	Uid int64 `json:"uid"` // 用户ID
	// Nickname string `json:"nickname"` // 昵称
	// 嵌入
	jwtcore.RegisteredClaims
}

func main() {
	r := gin.Default()

	accessKey := "access key"
	// 创建资源令牌管理器
	accessTM := jwtcore.NewTokenManager[Claims](accessKey, 10*time.Minute)
	// 创建认证中间件构建器
	builder := ujwt.NewMiddlewareBuilder[Claims](accessTM)

	// // 构建全局登录认证中间件
	// gloAuthMiddleware := builder.
	// 	// 忽略 "/login", "/signup" 这两个 full path 的认证。
	// 	IgnoreFullPath("/login", "/signup").Build()
	// // 全局拦截
	// r.Use(gloAuthMiddleware)

	// 单独拦截
	r.GET("/profile", builder.Build(), func(c *gin.Context) {
		// 获取 claims
		clm, ok := ujwt.ClaimsFromContext[Claims](c.Request.Context())
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
	// 创建刷新令牌的管理器
	refreshTM := jwtcore.NewTokenManager[Claims](refreshKey, 24*time.Hour)
	// 创建刷新令牌函数的构建器
	refreshManager := ujwt.NewRefreshManager[Claims](accessTM, refreshTM,
		ujwt.WithRotateRefreshToken[Claims](true), // 允许轮换 refresh token
	)

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

	// 内部已经开启了 refresh 令牌的认证,因此可以直接使用 refreshManager.Handler 与 relativePath 绑定
	r.POST("/refresh-token", refreshManager.Handler)

	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}

```



# `limit` package

该`limit`包为 gin 提供了限流中间件，使您快速完成全局的限流或者针对 IP 的限流。

在 [limiter](https://github.com/udugong/limiter) 仓库中提供了 `Limiter` 接口的实现。

- [滑动窗口限流](#滑动窗口限流)
- [活跃请求数限流](#活跃请求数限流)
- [桶限流](#桶限流)



## 滑动窗口限流

```go
package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	limit "github.com/udugong/ginx/middlewares/ratelimit/slidewindowlimit"
	"github.com/udugong/limiter/slidewindowlimit"
)

func main() {
	rdb := InitRedis()
	// github.com/udugong/limiter 中提供了一些 Limiter 接口的实现
	// 这里使用 slidewindowlimit 创建一个基于 redis, 1000/s 的滑动窗口限流器
	limiter := slidewindowlimit.NewRedisSlidingWindowLimiter(rdb, time.Second, 1000)

	// 本地滑动窗口限流 import "github.com/udugong/ukit/queue"
	// q := queue.NewCircularQueue[time.Time](1000)
	// limiter := slidewindowlimit.NewLocalSlideWindowLimiter(time.Second, q)

	builder := limit.NewBuilder(limiter)

	r := gin.Default()
	// 默认是全局限流
	// 控制所有的请求的速率 最多访问 1000次
	// r.Use(builder.Build())

	// 根据 IP 限流
	// 每个 IP 每秒 最多访问 1000次
	routes := r.Use(builder.SetKeyGenFuncByIP().Build())
	routes.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.Run()
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



## 活跃请求数限流

```go
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	limit "github.com/udugong/ginx/middlewares/ratelimit/activelimit"
	"github.com/udugong/limiter/activelimit"
)

func main() {
	rdb := InitRedis()
	// github.com/udugong/limiter 中提供了一些 Limiter 接口的实现
	// 这里使用 activelimit 创建一个基于 redis, 最大活跃请求数为 10 的活跃请求数限流器
	limiter := activelimit.NewRedisActiveLimiter(rdb, 10)

	// 本地活跃请求数限流
	// limiter := activelimit.NewLocalActiveLimiter(10)

	builder := limit.NewBuilder(limiter)

	r := gin.Default()
	// 默认是全局限流
	// 控制所有的活跃请求 最多有 10 个请求
	// r.Use(builder.Build())

	// 根据 IP 限流
	// 每个 IP 每最多 10 个活跃请求
	routes := r.Use(builder.SetKeyGenFuncByIP().Build())
	routes.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.Run()
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



## 桶限流

在初始化时可以使用 [limiter](https://github.com/udugong/limiter) 仓库提供的漏桶限流或者令牌桶限流。

- 无令牌时阻塞
- 无令牌时返回

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	limit "github.com/udugong/ginx/middlewares/ratelimit/bucketlimit"
	"github.com/udugong/limiter/bucketlimit"
)

func main() {
	// github.com/udugong/limiter 中提供了一些 Limiter 接口的实现
	// 这里使用 bucketlimit 创建一个漏桶限流
	limiter := bucketlimit.NewLeakyBucketLimiter(time.Second)

	// 令牌桶限流: 每秒生成一个令牌,桶内最多3枚令牌
	// limiter := bucketlimit.NewTokenBucketLimiter(time.Second, 3)

	builder := limit.NewBuilder(limiter)

	r := gin.Default()

	// 无令牌时返回,没有令牌时直接返回 429
	routes := r.Use(builder.Build())

	// 无令牌时阻塞,直到超时
	// routes := r.Use(builder.BuildBlock())

	routes.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.Run()
}

```

