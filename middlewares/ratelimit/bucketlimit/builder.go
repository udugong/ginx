package activelimit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	limiter Limiter
	logger  *slog.Logger
}

// NewBuilder 创建一个 Builder
func NewBuilder(limiter Limiter) *Builder {
	go limiter.Put()
	return &Builder{
		limiter: limiter,
		logger:  slog.Default(),
	}
}

func (b *Builder) SetLogger(logger *slog.Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		switch {
		case err == nil:
			if limited {
				ctx.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			ctx.Next()
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			ctx.AbortWithStatus(http.StatusGatewayTimeout)
			return
		default:
			b.logger.LogAttrs(ctx.Request.Context(), slog.LevelError,
				"限流器出现错误", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	return b.limiter.Limit(ctx.Request.Context(), "")
}

func (b *Builder) BuildBlock() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.blockLimit(ctx)
		switch {
		case err == nil:
			if limited {
				ctx.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			ctx.Next()
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			ctx.AbortWithStatus(http.StatusGatewayTimeout)
			return
		default:
			b.logger.LogAttrs(ctx.Request.Context(), slog.LevelError,
				"限流器出现错误", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}

func (b *Builder) blockLimit(ctx *gin.Context) (bool, error) {
	return b.limiter.BlockLimit(ctx.Request.Context(), "")
}
