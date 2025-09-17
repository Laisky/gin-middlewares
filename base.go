// Package middlewares some useful middlewares for gin
package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ctxkey string

const (
	// CtxKeyGin key of gin ctx that saved in request.context
	CtxKeyGin  ctxkey = "@laisky-gmw:gin"
	CtxKeyLock ctxkey = "@laisky-gmw:lock"
)

// FromStd convert std handler to gin.Handler, with gin context embedded
func FromStd(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx == nil || ctx.Request == nil {
			panic("gin context or request is nil in FromStd")
		}
		r2 := ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), CtxKeyGin, ctx))
		handler(ctx.Writer, r2)
	}
}

// GetGinCtxFromStdCtx get gin context from standard request.context by GinCtxKey
func GetGinCtxFromStdCtx(ctx context.Context) (*gin.Context, bool) {
	if ctx == nil {
		Logger.Warn("context is nil, cannot get gin context")
		return nil, false
	}

	if gctx, ok := ctx.(*gin.Context); ok && gctx != nil {
		return gctx, true
	}

	if v := ctx.Value(CtxKeyGin); v != nil {
		if gctx, ok := v.(*gin.Context); ok && gctx != nil {
			return gctx, true
		}
	}

	return nil, false
}
