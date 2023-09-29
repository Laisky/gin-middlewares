package middlewares

import (
	gutils "github.com/Laisky/go-utils/v4"
	"github.com/gin-gonic/gin"
)

// TraceID get trace id from context
func TraceID(ctx *gin.Context) string {
	val := ctx.GetString(defaultCtxKeyTraceID)
	if val == "" {
		val = ctx.Request.Header.Get(defaultCtxKeyTraceID)
	}

	if val == "" {
		val = gutils.UUID1()
	}

	ctx.Set(defaultCtxKeyTraceID, val)
	return val
}

// SpanID get span id from context
func SpanID(ctx *gin.Context) string {
	val := ctx.GetString(defaultCtxKeySpanID)
	if val == "" {
		val = ctx.Request.Header.Get(defaultCtxKeySpanID)
	}

	if val == "" {
		val = gutils.UUID1()
	}

	ctx.Set(defaultCtxKeySpanID, val)
	return val
}
