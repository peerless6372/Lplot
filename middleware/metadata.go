package middleware

import (
	"github.com/peerless6372/Lplot/utils/metadata"
	"lzh/gin-gonic/gin"
)

func Metadata() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		UseMetadata(ctx)
		ctx.Next()
	}
}

func UseMetadata(ctx *gin.Context) {
	if _, ok := metadata.CtxFromGinContext(ctx); !ok {
		metadata.GinCtxWithCtx(ctx, metadata.NewContext4Gin())
	}
}
