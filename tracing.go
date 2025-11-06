package middlewares

import (
	"github.com/Laisky/errors/v2"
	gutils "github.com/Laisky/go-utils/v6"
	"github.com/gin-gonic/gin"
)

// TraceID get trace id from context
func TraceID(ctx *gin.Context) (gutils.JaegerTracingID, error) {
	var val string
	if ctx != nil {
		// prefer value set in gin context
		val = ctx.GetString(gutils.TracingKey)
		// then try request header when request is available
		if val == "" && ctx.Request != nil {
			val = ctx.Request.Header.Get(gutils.TracingKey)
		}
	} else {
		Logger.Warn("TraceID called with nil gin.Context; creating new trace id")
	}

	if val == "" {
		if tid, err := gutils.NewJaegerTracingID(0, 0, 0, 0); err != nil {
			return tid, errors.Wrap(err, "new trace id")
		} else {
			val = tid.String()
		}
	}

	if ctx != nil {
		ctx.Set(gutils.TracingKey, val)
	} else {
		Logger.Warn("TraceID generated but cannot embed into nil gin.Context")
	}
	return gutils.JaegerTracingID(val), nil
}
