package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"time"

	gutils "github.com/Laisky/go-utils/v6"
	glog "github.com/Laisky/go-utils/v6/log"
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

// Ctx get request context from gin.Context
func Ctx(c *gin.Context) context.Context {
	// base context prefers request context when available
	var base context.Context
	if c != nil && c.Request != nil && c.Request.Context() != nil {
		base = c.Request.Context()
	} else {
		Logger.Warn("gin request or request context is nil, using background context in Ctx()")
		base = context.Background()
	}
	ctx := SetLogger(base, GetLogger(c))
	if c != nil {
		ctx = context.WithValue(ctx, CtxKeyGin, c)
	}

	if tid, err := TraceID(c); err != nil {
		GetLogger(ctx).Error("failed to get traceID", zap.Error(err))
	} else {
		ctx = context.WithValue(ctx, gutils.TracingKey, tid)
	}

	return ctx
}

// BackgroundCtx get background context from gin.Context
func BackgroundCtx(c *gin.Context) context.Context {
	if c == nil {
		panic("gin context is nil in BackgroundCtx")
	}

	ctx := SetLogger(c, GetLogger(c))
	ctx = context.WithValue(ctx, CtxKeyGin, c)

	if tid, err := TraceID(c); err != nil {
		GetLogger(ctx).Error("failed to get traceID", zap.Error(err))
	} else {
		ctx = context.WithValue(ctx, gutils.TracingKey, tid)
	}

	return ctx
}

// NewLoggerMiddleware middleware to logging
func NewLoggerMiddleware(optfs ...LoggerMwOptFunc) gin.HandlerFunc {
	opt := new(loggerMwOpt).fillDefault().applyOpts(optfs...)
	return func(ctx *gin.Context) {
		startAt := gutils.Clock.GetUTCNow()

		var traceID string
		if tid, err := TraceID(ctx); err != nil {
			opt.logger.Error("failed to get traceID", zap.Error(err))
		} else {
			traceID = tid.String()
		}

		// get logger
		logger := opt.logger
		if ctx == nil {
			// This should never happen in normal gin flow, but warn just in case.
			Logger.Warn("NewLoggerMiddleware received nil gin.Context; skip ctx-bound operations")
		}
		if ctx != nil {
			if loggeri, ok := ctx.Get(opt.ctxKeyLogger); ok {
				if l, ok := loggeri.(glog.Logger); ok && l != nil {
					logger = l
				}
			}
		}
		var urlStr, remote, host string
		if ctx != nil && ctx.Request != nil {
			if ctx.Request.URL != nil {
				urlStr = ctx.Request.URL.String()
			}
			remote = ctx.Request.RemoteAddr
			host = ctx.Request.Host
		} else if ctx != nil && ctx.Request == nil {
			Logger.Warn("gin request is nil in NewLoggerMiddleware; url/remote/host unavailable")
		}
		logger = logger.With(
			zap.String("url", urlStr),
			zap.String("remote", remote),
			zap.String("host", host),
			zap.String("trace_id", traceID),
			zap.String("cost", gutils.CostSecs(time.Since(startAt))),
		)

		// only log request size when method is not GET/HEAD/OPTIONS
		if ctx != nil && ctx.Request != nil && !gutils.Contains([]string{
			http.MethodHead, http.MethodGet, http.MethodOptions,
		}, ctx.Request.Method) {
			logger = logger.With(
				zap.String("request_size",
					gutils.HumanReadableByteCount(ctx.Request.ContentLength, true)),
			)
		} else if ctx != nil && ctx.Request == nil {
			Logger.Warn("gin request is nil; cannot log request size")
		}

		if ctx != nil {
			SetLogger(ctx, logger)
			ctx.Header(gutils.TracingKey, traceID)
			ctx.Next()
		} else {
			Logger.Warn("gin context is nil; cannot set logger into context or proceed Next()")
		}

		if ctx != nil && ctx.Writer != nil {
			logger = logger.With(zap.String("response_size",
				gutils.HumanReadableByteCount(int64(ctx.Writer.Size()), true)))
		} else if ctx != nil && ctx.Writer == nil {
			Logger.Warn("gin writer is nil; cannot log response size")
		}
		var status string
		if opt.colored {
			status = coloredStatus(ctx)
		} else {
			if ctx != nil && ctx.Writer != nil && ctx.Request != nil {
				status = strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method
			} else {
				Logger.Warn("missing writer or request; cannot compose status line")
			}
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
	if ctx == nil || ctx.Writer == nil || ctx.Request == nil {
		return ""
	}
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
	if ctx == nil {
		return glog.Shared.Named("gin")
	}
	switch c := ctx.(type) {
	case *gin.Context:
		if c != nil {
			if loggeri, ok := c.Get(defaultCtxKeyLogger); ok && loggeri != nil {
				if logger, ok := loggeri.(glog.Logger); ok && logger != nil {
					return logger
				}
			}
		}
		// c is a *gin.Context (possibly nil). Do not call ctx.Value because it would
		// dispatch to (*gin.Context).Value on a nil receiver and panic. Fall through
		// to return the default logger below.
	default:
		if loggeri := ctx.Value(defaultCtxKeyLogger); loggeri != nil {
			if logger, ok := loggeri.(glog.Logger); ok && logger != nil {
				return logger
			}
		}
	}

	return glog.Shared.Named("gin")
}

// SetLogger set logger into context
func SetLogger(ctx context.Context, logger glog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if gctx, ok := ctx.(*gin.Context); ok && gctx != nil {
		gctx.Set(defaultCtxKeyLogger, logger)
		if gctx.Request != nil && gctx.Request.Context() != nil {
			ctx = gctx.Request.Context()
		}
	}

	ctx = context.WithValue(ctx, defaultCtxKeyLogger, logger)
	return ctx
}
