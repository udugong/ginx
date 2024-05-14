package limit

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	limiter  Limiter
	genKeyFn func(ctx *gin.Context) string
	logger   *slog.Logger
}

// NewBuilder
// genKeyFn: 默认使用 IP 限流.
func NewBuilder(limiter Limiter) *Builder {
	return &Builder{
		limiter: limiter,
		genKeyFn: func(ctx *gin.Context) string {
			var b strings.Builder
			ip := ctx.ClientIP()
			b.Grow(11 + len(ip))
			b.WriteString("ip-limiter:")
			b.WriteString(ip)
			return b.String()
		},
		logger: slog.Default(),
	}
}

func (b *Builder) SetKeyGenFunc(fn func(*gin.Context) string) *Builder {
	b.genKeyFn = fn
	return b
}

func (b *Builder) SetLogger(logger *slog.Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			b.logger.LogAttrs(ctx.Request.Context(), slog.LevelError,
				"限流器出现错误", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	return b.limiter.Limit(ctx, b.genKeyFn(ctx))
}
