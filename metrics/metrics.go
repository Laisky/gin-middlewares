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
	defaultAddr = "127.0.0.1:8080"
	// defaultMetricPath      = "/metrics"
	defaultPProfPath = "/pprof"
	defaultGraceWait = 1 * time.Second
)

// option metric option argument
type option struct {
	addr, pprofPath string
	graceWait       time.Duration
}

// newOption create new default option
func newOption() *option {
	return &option{
		addr:      defaultAddr,
		pprofPath: defaultPProfPath,
		graceWait: defaultGraceWait,
	}
}

// OptFunc option of metrics
type OptFunc func(*option) error

// WithAddr set option addr
func WithAddr(addr string) OptFunc {
	return func(opt *option) error {
		opt.addr = addr
		return nil
	}
}

// WithGraceWait set wating time after graceful shutdown
func WithGraceWait(wait time.Duration) OptFunc {
	return func(opt *option) error {
		opt.graceWait = wait
		return nil
	}
}

// WithPprofPath set option pprofPath
func WithPprofPath(path string) OptFunc {
	return func(opt *option) error {
		opt.pprofPath = path
		return nil
	}
}

// Enable enable metrics for exsits gin server
func Enable(srv *gin.Engine, options ...OptFunc) (err error) {
	opt := newOption()
	for _, optf := range options {
		if err = optf(opt); err != nil {
			return errors.Wrap(err, "set option")
		}
	}

	pprof.Register(srv, opt.pprofPath)
	BindPrometheus(srv)
	return nil
}

// NewHTTPSrv start new gin server with metrics api
func NewHTTPSrv(ctx context.Context, options ...OptFunc) (srv *http.Server, err error) {
	opt := newOption()
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

	if err = Enable(router, options...); err != nil {
		return nil, errors.Wrap(err, "enable metric")
	}

	return
}

// BindPrometheus bind prometheus endpoint.
func BindPrometheus(s *gin.Engine) {
	p := ginprometheus.NewPrometheus("gin")
	p.Use(s)
}
