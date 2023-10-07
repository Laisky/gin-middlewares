package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"time"

	gutils "github.com/Laisky/go-utils/v4"
	glog "github.com/Laisky/go-utils/v4/log"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
)

const (
	defaultCtxKeyLogger = "gmw-logger"
)

// LoggerInterface logger interface
// type LoggerInterface interface {
// 	Debug(msg string, fields ...zapcore.Field)
// 	Info(msg string, fields ...zapcore.Field)
// }

type loggerMwOpt struct {
	logger                       glog.Logger
	colored                      bool
	ctxKeyLogger, ctxKeyTraceKey string
	level                        string
}

func (o *loggerMwOpt) applyOpts(optfs ...LoggerMwOptFunc) *loggerMwOpt {
	for _, optf := range optfs {
		optf(o)
	}

	return o
}

func (o *loggerMwOpt) fillDefault() *loggerMwOpt {
	o.logger = Logger.Named("gin-middlewares")
	o.level = glog.LevelDebug.String()
	o.ctxKeyLogger = defaultCtxKeyLogger
	o.ctxKeyTraceKey = gutils.TracingKey
	return o
}

// LoggerMwOptFunc logger options
type LoggerMwOptFunc func(opt *loggerMwOpt)

// WithLoggerMwColored enable coloered log
func WithLoggerMwColored() LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.colored = true
	}
}

// WithLoggerCtxKey embedded logger into context
// func WithLoggerCtxKey(key string) LoggerMwOptFunc {
// 	return func(opt *loggerMwOpt) {
// 		opt.ctxKeyLogger = key
// 	}
// }

// WithTracingCtxKey embedded traceID into context
func WithTracingCtxKey(key string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.ctxKeyTraceKey = key
	}
}

// WithLevel (optional) set log level
//
// only support debug/info
//
// default to debug
func WithLevel(level string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.level = level
	}
}

// WithLogger set default logger
func WithLogger(logger glog.Logger) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.logger = logger
	}
}

// NewLoggerMiddleware middleware to logging
func NewLoggerMiddleware(optfs ...LoggerMwOptFunc) gin.HandlerFunc {
	opt := new(loggerMwOpt).fillDefault().applyOpts(optfs...)
	return func(ctx *gin.Context) {
		startAt := gutils.Clock.GetUTCNow()

		var traceID string
		if tid, err := TraceID(ctx); err != nil {
			traceID = tid.String()
		}

		// get logger
		logger := opt.logger
		if loggeri, ok := ctx.Get(opt.ctxKeyLogger); ok {
			if l, ok := loggeri.(glog.Logger); ok && l != nil {
				logger = l
			}
		}
		logger = logger.With(
			zap.String("url", ctx.Request.URL.String()),
			zap.String("remote", ctx.Request.RemoteAddr),
			zap.String("host", ctx.Request.Host),
			zap.String("trace_id", traceID),
			zap.String("cost", gutils.CostSecs(time.Since(startAt))),
		)

		// only log request size when method is not GET/HEAD/OPTIONS
		if !gutils.Contains([]string{
			http.MethodHead, http.MethodGet, http.MethodOptions,
		}, ctx.Request.Method) {
			logger = logger.With(
				zap.String("request_size",
					gutils.HumanReadableByteCount(ctx.Request.ContentLength, true)),
			)
		}

		SetLogger(ctx, logger)
		ctx.Header(gutils.TracingKey, traceID)
		ctx.Next()

		logger = logger.With(zap.String("response_size",
			gutils.HumanReadableByteCount(int64(ctx.Writer.Size()), true)))
		var status string
		if opt.colored {
			status = coloredStatus(ctx)
		} else {
			status = strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method
		}

		switch opt.level {
		case string(glog.LevelInfo):
			logger.Info(status)
		default:
			logger.Debug(status)
		}
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

// GetLogger get logger from context
func GetLogger(ctx context.Context) (logger glog.Logger) {
	if gctx, ok := ctx.(*gin.Context); ok && gctx != nil {
		if loggeri, ok := gctx.Get(defaultCtxKeyLogger); ok && loggeri != nil {
			if logger, ok := loggeri.(glog.Logger); ok && logger != nil {
				return logger
			}
		}
	}

	if loggeri := ctx.Value(defaultCtxKeyLogger); loggeri != nil {
		if logger, ok := loggeri.(glog.Logger); ok && logger != nil {
			return logger
		}
	}

	return glog.Shared.Named("gin")
}

// SetLogger set logger into context
func SetLogger(ctx context.Context, logger glog.Logger) context.Context {
	if gctx, ok := ctx.(*gin.Context); ok && gctx != nil {
		gctx.Set(defaultCtxKeyLogger, logger)
		if gctx.Request != nil {
			ctx = gctx.Request.Context()
		}
	}

	ctx = context.WithValue(ctx, defaultCtxKeyLogger, logger)
	return ctx
}
