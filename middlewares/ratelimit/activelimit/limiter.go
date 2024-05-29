package activelimit

import "context"

// Limiter 活跃请求数限流
type Limiter interface {
	// Limit 有没有触发限流。key 就是限流对象
	// 活跃请求数增加1
	// bool 代表是否限流, true 就是要限流
	// error 限流器本身有没有错误
	Limit(ctx context.Context, key string) (bool, error)

	// Decr 活跃请求数减少1
	Decr(ctx context.Context, key string) error
}
