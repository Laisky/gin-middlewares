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
	defaultCtxKeyLogger gutils.CtxKey = "gmw-logger"
)

// LoggerInterface logger interface
// type LoggerInterface interface {
// 	Debug(msg string, fields ...zapcore.Field)
// 	Info(msg string, fields ...zapcore.Field)
// }

type loggerMwOpt struct {
	logger                       glog.Logger
	colored                      bool
	ctxKeyLogger, ctxKeyTraceKey gutils.CtxKey
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
		opt.ctxKeyTraceKey = gutils.CtxKey(key)
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
		traceID := extractTraceIDForLogger(ctx, opt.logger)
		logger := resolveRequestLogger(ctx, opt)
		logger = withRequestFields(logger, ctx, traceID, startAt)
		logger = withRequestSizeField(logger, ctx)
		advanceRequest(ctx, logger, traceID)
		logger = withResponseSizeField(logger, ctx)
		logByLevel(logger, buildStatus(ctx, opt.colored), opt.level)
	}
}

// extractTraceIDForLogger gets trace id from gin context for middleware logging.
// The ctx parameter is the current gin request context.
// The fallbackLogger parameter records trace extraction errors.
// It returns the trace id string if available; otherwise returns an empty string.
func extractTraceIDForLogger(ctx *gin.Context, fallbackLogger glog.Logger) string {
	if tid, err := TraceID(ctx); err != nil {
		fallbackLogger.Error("failed to get traceID", zap.Error(err))
		return ""
	} else {
		return tid.String()
	}
}

// resolveRequestLogger resolves logger from gin context and falls back to middleware default logger.
// The ctx parameter is the current gin request context.
// The opt parameter contains middleware logger options and context keys.
// It returns a non-nil logger for this request.
func resolveRequestLogger(ctx *gin.Context, opt *loggerMwOpt) glog.Logger {
	logger := opt.logger
	if ctx == nil {
		// This should never happen in normal gin flow, but warn just in case.
		Logger.Warn("NewLoggerMiddleware received nil gin.Context; skip ctx-bound operations")
		return logger
	}

	if loggeri, ok := ctx.Get(opt.ctxKeyLogger); ok {
		if l, ok := loggeri.(glog.Logger); ok && l != nil {
			return l
		}
	}

	return logger
}

// withRequestFields appends request-related metadata to logger fields.
// The logger parameter is the current request logger.
// The ctx parameter is the current gin request context.
// The traceID parameter is the extracted trace id string.
// The startAt parameter is the request start timestamp used to compute latency.
// It returns a logger decorated with request metadata fields.
func withRequestFields(logger glog.Logger, ctx *gin.Context, traceID string, startAt time.Time) glog.Logger {
	urlStr, remote, host := requestBasicFields(ctx)
	return logger.With(
		zap.String("url", urlStr),
		zap.String("remote", remote),
		zap.String("host", host),
		zap.String("trace_id", traceID),
		zap.String("cost", gutils.CostSecs(time.Since(startAt))),
	)
}

// requestBasicFields extracts url/remote/host from request context when possible.
// The ctx parameter is the current gin request context.
// It returns url string, remote address, and host string in that order.
func requestBasicFields(ctx *gin.Context) (string, string, string) {
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

	return urlStr, remote, host
}

// withRequestSizeField appends request payload size when request method carries body.
// The logger parameter is the current request logger.
// The ctx parameter is the current gin request context.
// It returns logger with request_size field when applicable.
func withRequestSizeField(logger glog.Logger, ctx *gin.Context) glog.Logger {
	// only log request size when method is not GET/HEAD/OPTIONS
	if ctx != nil && ctx.Request != nil && !gutils.Contains([]string{
		http.MethodHead, http.MethodGet, http.MethodOptions,
	}, ctx.Request.Method) {
		return logger.With(
			zap.String("request_size",
				gutils.HumanReadableByteCount(ctx.Request.ContentLength, true)),
		)
	} else if ctx != nil && ctx.Request == nil {
		Logger.Warn("gin request is nil; cannot log request size")
	}

	return logger
}

// advanceRequest writes logger and trace id back to gin context and continues middleware chain.
// The ctx parameter is the current gin request context.
// The logger parameter is the request logger that should be stored.
// The traceID parameter is the trace id string written to response headers.
// It does not return a value.
func advanceRequest(ctx *gin.Context, logger glog.Logger, traceID string) {
	if ctx != nil {
		SetLogger(ctx, logger)
		ctx.Header(gutils.TracingKey.String(), traceID)
		ctx.Next()
		return
	}

	Logger.Warn("gin context is nil; cannot set logger into context or proceed Next()")
}

// withResponseSizeField appends response payload size to logger fields when writer exists.
// The logger parameter is the current request logger.
// The ctx parameter is the current gin request context.
// It returns logger with response_size field when possible.
func withResponseSizeField(logger glog.Logger, ctx *gin.Context) glog.Logger {
	if ctx != nil && ctx.Writer != nil {
		return logger.With(zap.String("response_size",
			gutils.HumanReadableByteCount(int64(ctx.Writer.Size()), true)))
	} else if ctx != nil && ctx.Writer == nil {
		Logger.Warn("gin writer is nil; cannot log response size")
	}

	return logger
}

// buildStatus builds final status text for request logs.
// The ctx parameter is the current gin request context.
// The colored parameter controls whether status should include ANSI color.
// It returns a status string suitable for final Info/Debug logging.
func buildStatus(ctx *gin.Context, colored bool) string {
	if colored {
		return coloredStatus(ctx)
	}

	if ctx != nil && ctx.Writer != nil && ctx.Request != nil {
		return strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method
	}

	Logger.Warn("missing writer or request; cannot compose status line")
	return ""
}

// logByLevel emits final status line with selected log level.
// The logger parameter is the request logger.
// The status parameter is the final status message content.
// The level parameter controls whether to log with info or debug.
// It does not return a value.
func logByLevel(logger glog.Logger, status, level string) {
	switch level {
	case string(glog.LevelInfo):
		logger.Info(status)
	default:
		logger.Debug(status)
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
