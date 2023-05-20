package metadata

import (
	"context"

	""lzh/gin-gonic/gin""
)

const _CTX_KEY = "goframework/net/metadata.ctx"

func CtxFromGinContext(c *gin.Context) (context.Context, bool) {
	if c != nil {
		if v, ok := c.Get(_CTX_KEY); ok {
			res := v.(context.Context)
			return res, true
		}
	}
	return nil, false
}

func GinCtxWithCtx(c *gin.Context, ctx context.Context) {
	if c != nil {
		c.Set(_CTX_KEY, ctx)
	}
}

func NewContext4Gin() context.Context {
	md := MD(map[string]interface{}{
		Notice: make(map[string]interface{}),
	})
	ctx := NewContext(context.Background(), md)
	return ctx
}
