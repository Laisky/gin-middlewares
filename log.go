package middlewares

import (
	"strconv"
	"time"

	gutils "github.com/Laisky/go-utils/v2"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
)

type loggerMwOpt struct {
	logger       gutils.LoggerItf
	colored      bool
	ctxKeyLogger string
}

func (o *loggerMwOpt) applyOpts(optfs ...LoggerMwOptFunc) *loggerMwOpt {
	for _, optf := range optfs {
		optf(o)
	}

	return o
}

func (o *loggerMwOpt) fillDefault() *loggerMwOpt {
	o.logger = gutils.Logger.Named("gin-middlewares")
	return o
}

type LoggerMwOptFunc func(opt *loggerMwOpt)

func WithLoggerMwColored() LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.colored = true
	}
}

func WithLoggerCtxKey(key string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.ctxKeyLogger = key
	}
}

func WithLogger(logger gutils.LoggerItf) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.logger = logger
	}
}

// NewLoggerMiddleware middleware to logging
func NewLoggerMiddleware(optfs ...LoggerMwOptFunc) gin.HandlerFunc {
	opt := new(loggerMwOpt).fillDefault().applyOpts(optfs...)
	return func(ctx *gin.Context) {
		startAt := gutils.Clock.GetUTCNow()

		ctx.Next()

		var status string
		if opt.colored {
			status = coloredStatus(ctx)
		} else {
			status = strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method

		}

		logger := opt.logger
		if loggeri, ok := ctx.Get(opt.ctxKeyLogger); ok {
			if l, ok := loggeri.(gutils.LoggerItf); ok && l != nil {
				logger = l
			}
		}

		logger.Info(status,
			zap.String("url", ctx.Request.URL.String()),
			zap.String("remote", ctx.Request.RemoteAddr),
			zap.String("host", ctx.Request.Host),
			zap.Int("size", ctx.Writer.Size()),
			zap.Duration("latency_ms", time.Since(startAt)*1000),
		)
	}
}

// coloredStatus zap field 会做二次转译，导致 ANSI color 失效
func coloredStatus(ctx *gin.Context) string {
	codeStr := strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method
	switch ctx.Writer.Status() / 100 {
	case 2:
		codeStr = gutils.Color(gutils.ANSIColorFgGreen, codeStr)
	case 4:
		codeStr = gutils.Color(gutils.ANSIColorFgYellow, codeStr)
	case 5:
		codeStr = gutils.Color(gutils.ANSIColorFgRed, codeStr)
	default:
		codeStr = gutils.Color(gutils.ANSIColorFgCyan, codeStr)
	}

	return codeStr
}
