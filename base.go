// Package middlewares some useful middlewares for gin
package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ctxkey string

// GinCtxKey key of gin ctx that saved in request.context
var GinCtxKey ctxkey = "gin"

// FromStd convert std handler to gin.Handler, with gin context embedded
func FromStd(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r2 := ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), GinCtxKey, ctx))
		handler(ctx.Writer, r2)
	}
}

// GetGinCtxFromStdCtx get gin context from standard request.context by GinCtxKey
func GetGinCtxFromStdCtx(ctx context.Context) *gin.Context {
	return ctx.Value(GinCtxKey).(*gin.Context)
}
