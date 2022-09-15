package middlewares

import (
	"fmt"

	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	defaultCookiePath     = "/"
	defaultCookieSecure   = false
	defaultCookieHTTPOnly = false
)

type setCookieOption struct {
	cookieMaxAge                 int
	cookiePath, cookieHost       string
	cookieSecure, cookieHttpOnly bool
}

func (o *setCookieOption) fillDefault(ctx *gin.Context) *setCookieOption {
	o.cookiePath = defaultCookiePath
	o.cookieSecure = defaultCookieSecure
	o.cookieHttpOnly = defaultCookieHTTPOnly
	o.cookieHost = ctx.Request.Host

	if ctx.Request.URL.Port() != "" {
		o.cookieHost += ":" + ctx.Request.URL.Port()
	}

	return o
}

func (o *setCookieOption) applyOpts(opts ...SetCookieOption) (*setCookieOption, error) {
	for _, f := range opts {
		if err := f(o); err != nil {
			return nil, errors.Wrap(err, "apply auth options")
		}
	}

	return o, nil
}

// SetCookieOption auth cookie options
type SetCookieOption func(*setCookieOption) error

// WithCookieMaxAge set auth cookie's maxAge
func WithCookieMaxAge(maxAge int) SetCookieOption {
	return func(opt *setCookieOption) error {
		if maxAge < 0 {
			return fmt.Errorf("maxAge should not less than 0, got %v", maxAge)
		}

		opt.cookieMaxAge = maxAge
		return nil
	}
}

// WithCookiePath set auth cookie's path
func WithCookiePath(path string) SetCookieOption {
	Logger.Debug("set auth cookie path", zap.String("path", path))
	return func(opt *setCookieOption) error {
		opt.cookiePath = path
		return nil
	}
}

// WithCookieSecure set auth cookie's secure
func WithCookieSecure(secure bool) SetCookieOption {
	Logger.Debug("set auth cookie secure", zap.Bool("secure", secure))
	return func(opt *setCookieOption) error {
		opt.cookieSecure = secure
		return nil
	}
}

// WithCookieHTTPOnly set auth cookie's HTTPOnly
func WithCookieHTTPOnly(httpOnly bool) SetCookieOption {
	Logger.Debug("set auth cookie httpOnly", zap.Bool("httpOnly", httpOnly))
	return func(opt *setCookieOption) error {
		opt.cookieHttpOnly = httpOnly
		return nil
	}
}

// WithCookieHost set auth cookie's host
func WithCookieHost(host string) SetCookieOption {
	Logger.Debug("set auth cookie host", zap.String("host", host))
	return func(opt *setCookieOption) error {
		opt.cookieHost = host
		return nil
	}
}

// SetCookie set jwt token to cookies
func SetCookie(ctx *gin.Context,
	name, value string,
	opts ...SetCookieOption) (err error) {
	opt, err := new(setCookieOption).fillDefault(ctx).applyOpts()
	if err != nil {
		return err
	}

	ctx.SetCookie(name,
		value,
		opt.cookieMaxAge,
		opt.cookiePath,
		opt.cookieHost,
		opt.cookieSecure,
		opt.cookieHttpOnly)
	return nil
}
