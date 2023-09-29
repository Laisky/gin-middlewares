package middlewares

import (
	"net/http"
	"strconv"
	"time"

	gutils "github.com/Laisky/go-utils/v4"
	glog "github.com/Laisky/go-utils/v4/log"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
)

const (
	defaultCtxKeyLogger  = "gmw-logger"
	defaultCtxKeyTraceID = "uber-trace-id"
	defaultCtxKeySpanID  = "uber-span-id"
)

// LoggerInterface logger interface
// type LoggerInterface interface {
// 	Debug(msg string, fields ...zapcore.Field)
// 	Info(msg string, fields ...zapcore.Field)
// }

type loggerMwOpt struct {
	logger                                    glog.Logger
	colored                                   bool
	ctxKeyLogger, ctxKeyTraceID, ctxKeySpanID string
	level                                     string
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
	o.ctxKeySpanID = defaultCtxKeySpanID
	o.ctxKeyTraceID = defaultCtxKeyTraceID
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

// WithTraceIDCtxKey embedded traceID into context
func WithTraceIDCtxKey(key string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.ctxKeyTraceID = key
	}
}

// WithSpanIDCtxKey embedded spanID into context
func WithSpanIDCtxKey(key string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.ctxKeySpanID = key
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
			zap.String("trace_id", TraceID(ctx)),
			zap.String("span_id", SpanID(ctx)),
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
		ctx.Header(defaultCtxKeyTraceID, TraceID(ctx))
		ctx.Header(defaultCtxKeySpanID, SpanID(ctx))
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
func GetLogger(ctx *gin.Context) (logger glog.Logger) {
	if loggeri, ok := ctx.Get(defaultCtxKeyLogger); ok {
		if logger, ok := loggeri.(glog.Logger); ok && logger != nil {
			return logger
		}
	}

	return glog.Shared.Named("gin")
}

// SetLogger set logger into context
func SetLogger(ctx *gin.Context, logger glog.Logger) {
	ctx.Set(defaultCtxKeyLogger, logger)
}
