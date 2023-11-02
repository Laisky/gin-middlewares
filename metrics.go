package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/Laisky/errors/v2"
	ginprometheus "github.com/Laisky/go-gin-prometheus"
	"github.com/Laisky/zap"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var (
	defaultMetricAddr = "localhost:8080"
	// defaultMetricPath      = "/metrics"
	defaultPProfPath       = "/pprof"
	defaultMetricGraceWait = 1 * time.Second
)

// metricOption metric option argument
type metricOption struct {
	addr, pprofPath string
	graceWait       time.Duration
}

// newMetricOption create new default option
func newMetricOption() *metricOption {
	return &metricOption{
		addr:      defaultMetricAddr,
		pprofPath: defaultPProfPath,
		graceWait: defaultMetricGraceWait,
	}
}

// MetricsOptFunc option of metrics
type MetricsOptFunc func(*metricOption) error

// WithMetricAddr set option addr
func WithMetricAddr(addr string) MetricsOptFunc {
	return func(opt *metricOption) error {
		opt.addr = addr
		return nil
	}
}

// WithMetricGraceWait set wating time after graceful shutdown
func WithMetricGraceWait(wait time.Duration) MetricsOptFunc {
	return func(opt *metricOption) error {
		opt.graceWait = wait
		return nil
	}
}

// WithPprofPath set option pprofPath
func WithPprofPath(path string) MetricsOptFunc {
	return func(opt *metricOption) error {
		opt.pprofPath = path
		return nil
	}
}

// EnableMetric enable metrics for exsits gin server
func EnableMetric(srv gin.IRouter, options ...MetricsOptFunc) (err error) {
	opt := newMetricOption()
	for _, optf := range options {
		if err = optf(opt); err != nil {
			return errors.Wrap(err, "set option")
		}
	}

	pprof.Register(srv, opt.pprofPath)
	BindPrometheus(srv)
	return nil
}

// NewHTTPMetricSrv start new gin server with metrics api
func NewHTTPMetricSrv(ctx context.Context, options ...MetricsOptFunc) (srv *http.Server, err error) {
	opt := newMetricOption()
	for _, optf := range options {
		if err = optf(opt); err != nil {
			return nil, errors.Wrap(err, "set option")
		}
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	srv = &http.Server{
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              opt.addr,
		Handler:           router,
	}

	go func() {
		<-ctx.Done()
		Logger.Info("got signal to shutdown metric server")
		timingCtx, cancel := context.WithTimeout(context.Background(), opt.graceWait)
		defer cancel()
		if err := srv.Shutdown(timingCtx); err != nil {
			Logger.Error("shutdown metrics server", zap.Error(err), zap.String("addr", opt.addr))
		}
	}()

	if err = EnableMetric(router, options...); err != nil {
		return nil, errors.Wrap(err, "enable metric")
	}

	return
}

// BindPrometheus bind prometheus endpoint.
func BindPrometheus(s gin.IRouter) {
	p := ginprometheus.NewPrometheus("gin")
	p.Use(s)
}
