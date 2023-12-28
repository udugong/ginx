package jwt

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/udugong/ginx/jwt/jwtcore"
)

// RefreshHandlerBuilder 定义刷新令牌的 gin.HandlerFunc 构建器.
// accessTM: 资源令牌管理
// refreshTM: 刷新令牌管理
// rotateRefreshToken: 是否轮换刷新令牌。默认为 false
// exposeAccessHeader: 暴露到外部的资源令牌的请求头。默认为 x-access-token
// exposeRefreshHeader: 暴露到外部的刷新令牌的请求头。默认为 x-refresh-token
// refreshAuthHandler: 刷新令牌认证的处理函数。默认使用 refreshTM 作为参数创建的 MiddlewareBuilder 进行认证
type RefreshHandlerBuilder[T jwt.Claims, PT jwtcore.Claims[T]] struct {
	accessTM            jwtcore.TokenManager[T, PT]
	refreshTM           jwtcore.TokenManager[T, PT]
	rotateRefreshToken  bool   // 是否轮换刷新令牌
	exposeAccessHeader  string // 暴露到外部的资源请求头
	exposeRefreshHeader string // 暴露到外部的刷新请求头
	refreshAuthHandler  gin.HandlerFunc
}

// NewRefreshHandlerBuilder 创建一个刷新令牌的 gin.HandlerFunc 构建器.
func NewRefreshHandlerBuilder[T jwt.Claims, PT jwtcore.Claims[T]](
	accessTM jwtcore.TokenManager[T, PT], refreshTM jwtcore.TokenManager[T, PT],
	options ...Option[T, PT]) *RefreshHandlerBuilder[T, PT] {
	builder := &RefreshHandlerBuilder[T, PT]{
		accessTM:            accessTM,
		refreshTM:           refreshTM,
		rotateRefreshToken:  false,
		exposeAccessHeader:  "x-access-token",
		exposeRefreshHeader: "x-refresh-token",
		refreshAuthHandler:  NewMiddlewareBuilder[T, PT](refreshTM).Build(),
	}
	return builder.WithOptions(options...)
}

type Option[T jwt.Claims, PT jwtcore.Claims[T]] interface {
	apply(*RefreshHandlerBuilder[T, PT])
}

type optionFunc[T jwt.Claims, PT jwtcore.Claims[T]] func(*RefreshHandlerBuilder[T, PT])

func (f optionFunc[T, PT]) apply(rh *RefreshHandlerBuilder[T, PT]) {
	f(rh)
}

func WithRotateRefreshToken[T jwt.Claims, PT jwtcore.Claims[T]](isRotate bool) Option[T, PT] {
	return optionFunc[T, PT](func(rh *RefreshHandlerBuilder[T, PT]) {
		rh.rotateRefreshToken = isRotate
	})
}

func WithExposeAccessHeader[T jwt.Claims, PT jwtcore.Claims[T]](header string) Option[T, PT] {
	return optionFunc[T, PT](func(rh *RefreshHandlerBuilder[T, PT]) {
		rh.exposeAccessHeader = header
	})
}

func WithExposeRefreshHeader[T jwt.Claims, PT jwtcore.Claims[T]](header string) Option[T, PT] {
	return optionFunc[T, PT](func(rh *RefreshHandlerBuilder[T, PT]) {
		rh.exposeRefreshHeader = header
	})
}

// Build 构建 gin.HandlerFunc.
func (rh *RefreshHandlerBuilder[T, PT]) Build(c *gin.Context) {
	rh.refreshAuthHandler(c)
	if c.IsAborted() {
		return
	}
	v, ok := c.Get(claimsKey)
	clm, ok := v.(T)
	if !ok {
		// 不应该命中该分支
		c.Status(http.StatusInternalServerError)
		log.Println("未知错误")
		return
	}

	accessToken, err := rh.accessTM.GenerateToken(clm)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Printf("生成 access token 失败; err: %v", err)
		return
	}

	c.Header(rh.exposeAccessHeader, accessToken)

	// 轮换刷新令牌
	if rh.rotateRefreshToken {
		refreshToken, err := rh.refreshTM.GenerateToken(clm)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("生成 refresh token 失败; err: %v", err)
			return
		}
		c.Header(rh.exposeRefreshHeader, refreshToken)
	}
	c.Status(http.StatusNoContent)
}

func (rh *RefreshHandlerBuilder[T, PT]) WithOptions(opts ...Option[T, PT]) *RefreshHandlerBuilder[T, PT] {
	c := rh.clone()
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func (rh *RefreshHandlerBuilder[T, PT]) clone() *RefreshHandlerBuilder[T, PT] {
	copyHandler := *rh
	return &copyHandler
}
