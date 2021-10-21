package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

var (
	defaultMetricAddr = "127.0.0.1:8080"
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
func EnableMetric(srv *gin.Engine, options ...MetricsOptFunc) (err error) {
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
		Addr:    opt.addr,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		log.Println("got signal to shutdown metric server")
		timingCtx, cancel := context.WithTimeout(context.Background(), opt.graceWait)
		defer cancel()
		if err := srv.Shutdown(timingCtx); err != nil {
			log.Println("shutdown metrics server", err)
		}
	}()

	if err = EnableMetric(router, options...); err != nil {
		return nil, errors.Wrap(err, "enable metric")
	}

	return
}

// BindPrometheus bind prometheus endpoint.
func BindPrometheus(s *gin.Engine) {
	p := ginprometheus.NewPrometheus("gin")
	p.Use(s)
}