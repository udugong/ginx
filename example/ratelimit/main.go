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
