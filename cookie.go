package middlewares

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
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
	if ctx != nil && ctx.Request != nil {
		o.cookieHost = ctx.Request.Host
	}

	return o
}

func (o *setCookieOption) applyOpts(opts ...SetCookieOption) (*setCookieOption, error) {
	for _, f := range opts {
		if err := f(o); err != nil {
			return nil, errors.Wrap(err, "apply cookie options")
		}
	}

	return o, nil
}

// normalizeCookieHost normalizes cookie domain input by stripping ports and invalid host parts.
// The host parameter accepts request host values or user-provided cookie host overrides.
// The returned host is safe for the Set-Cookie Domain attribute, or empty if no valid domain can be derived.
func normalizeCookieHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	if strings.Contains(host, "://") {
		if parsedURL, err := url.Parse(host); err == nil {
			host = parsedURL.Host
		}
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	} else if strings.Count(host, ":") == 1 {
		if hostPart, portPart, ok := strings.Cut(host, ":"); ok {
			if _, err := strconv.Atoi(portPart); err == nil {
				host = hostPart
			}
		}
	}

	host = strings.TrimSpace(strings.Trim(host, "[]"))
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return ""
	}

	// RFC 6265 domain attribute is host-name based; IPv6 literals are not valid domains.
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		return ""
	}

	if strings.Contains(host, ":") {
		return ""
	}

	return host
}

// SetCookieOption auth cookie options
type SetCookieOption func(*setCookieOption) error

// WithCookieMaxAge set auth cookie's maxAge
func WithCookieMaxAge(maxAge int) SetCookieOption {
	return func(opt *setCookieOption) error {
		if maxAge < 0 {
			return errors.Errorf("maxAge should not less than 0, got %v", maxAge)
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
	logger := GetLogger(ctx)
	opt, err := new(setCookieOption).fillDefault(ctx).applyOpts(opts...)
	if err != nil {
		logger.Debug("failed to apply cookie options",
			zap.String("cookie_name", name),
			zap.Error(err),
		)
		return errors.Wrap(err, "apply cookie options")
	}

	if ctx == nil {
		logger.Warn("SetCookie got nil gin.Context")
		return errors.New("gin context is nil")
	}
	if ctx.Writer == nil {
		logger.Warn("SetCookie got nil gin writer")
		return errors.New("gin writer is nil")
	}

	rawCookieHost := opt.cookieHost
	opt.cookieHost = normalizeCookieHost(opt.cookieHost)
	if rawCookieHost != opt.cookieHost {
		logger.Debug("normalized cookie host",
			zap.String("raw_cookie_host", rawCookieHost),
			zap.String("normalized_cookie_host", opt.cookieHost),
		)
	}

	logger.Debug("set cookie",
		zap.String("cookie_name", name),
		zap.Int("cookie_max_age", opt.cookieMaxAge),
		zap.String("cookie_path", opt.cookiePath),
		zap.String("cookie_domain", opt.cookieHost),
		zap.Bool("cookie_secure", opt.cookieSecure),
		zap.Bool("cookie_http_only", opt.cookieHttpOnly),
	)

	ctx.SetCookie(name,
		value,
		opt.cookieMaxAge,
		opt.cookiePath,
		opt.cookieHost,
		opt.cookieSecure,
		opt.cookieHttpOnly)
	return nil
}
