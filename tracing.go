package middlewares

import (
	"github.com/Laisky/errors/v2"
	gutils "github.com/Laisky/go-utils/v4"
	"github.com/gin-gonic/gin"
)

// TraceID get trace id from context
func TraceID(ctx *gin.Context) (gutils.JaegerTracingID, error) {
	val := ctx.GetString(gutils.TracingKey)
	if val == "" {
		val = ctx.Request.Header.Get(gutils.TracingKey)
	}

	if val == "" {
		if tid, err := gutils.NewJaegerTracingID(0, 0, 0, 0); err != nil {
			return tid, errors.Wrap(err, "new trace id")
		} else {
			val = tid.String()
		}
	}

	ctx.Set(gutils.TracingKey, val)
	return gutils.JaegerTracingID(val), nil
}
