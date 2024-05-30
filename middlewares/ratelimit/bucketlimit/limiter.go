package activelimit

import "context"

// Limiter 桶限流
type Limiter interface {
	// Put 往桶里放置。该方法需要异步执行 go Put()
	Put()

	// Close 关闭 Put() 方法
	Close()

	// Limit 有没有触发限流。
	// bool 代表是否限流, true 就是要限流, 若 Context.Err() != nil 也会返回 error
	// error 当调用了 Close() 时返回错误
	Limit(ctx context.Context, _ string) (bool, error)

	// BlockLimit 限流时阻塞直到超时。
	// bool 代表是否限流, true 就是要限流, 同时会返回 Context.Err()
	// error 当调用了 Close() 时返回错误
	BlockLimit(ctx context.Context, _ string) (bool, error)
}
