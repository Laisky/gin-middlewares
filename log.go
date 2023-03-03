package middlewares

import (
	"strconv"
	"time"

	gutils "github.com/Laisky/go-utils/v4"
	glog "github.com/Laisky/go-utils/v4/log"
	"github.com/Laisky/zap"
	"github.com/Laisky/zap/zapcore"
	"github.com/gin-gonic/gin"
)

type loggerMwOpt struct {
	logger       glog.Logger
	colored      bool
	ctxKeyLogger string
	level        glog.Level
}

func (o *loggerMwOpt) applyOpts(optfs ...LoggerMwOptFunc) *loggerMwOpt {
	for _, optf := range optfs {
		optf(o)
	}

	return o
}

func (o *loggerMwOpt) fillDefault() *loggerMwOpt {
	o.logger = Logger.Named("gin-middlewares")
	o.level = glog.LevelDebug
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
func WithLoggerCtxKey(key string) LoggerMwOptFunc {
	return func(opt *loggerMwOpt) {
		opt.ctxKeyLogger = key
	}
}

// WithLevel (optional) set log level
//
// only support debug/info
//
// default to debug
func WithLevel(level glog.Level) LoggerMwOptFunc {
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

		ctx.Next()

		var status string
		if opt.colored {
			status = coloredStatus(ctx)
		} else {
			status = strconv.Itoa(ctx.Writer.Status()) + " " + ctx.Request.Method

		}

		logger := opt.logger
		if loggeri, ok := ctx.Get(opt.ctxKeyLogger); ok {
			if l, ok := loggeri.(glog.Logger); ok && l != nil {
				logger = l
			}
		}

		fields := []zapcore.Field{
			zap.String("url", ctx.Request.URL.String()),
			zap.String("remote", ctx.Request.RemoteAddr),
			zap.String("host", ctx.Request.Host),
			zap.Int("size", ctx.Writer.Size()),
			zap.String("cost", gutils.CostSecs(time.Since(startAt))),
		}
		switch opt.level {
		case glog.LevelInfo:
			logger.Info(status, fields...)
		default:
			logger.Debug(status, fields...)
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
