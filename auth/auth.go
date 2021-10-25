package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Laisky/gin-middlewares/library"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"
)

const (
	// defaultTokenName jwt token cookie name
	defaultTokenName = "token"
	// defaultUserIDCtxKey key of user ID in jwt token
	// defaultUserIDCtxKey           = "auth_uid"
	defaultJWTTokenExpireDuration = 7 * 24 * time.Hour

	defaultCookiePath     = "/"
	defaultCookieSecure   = false
	defaultCookieHTTPOnly = false
)

// OptFunc auth option
type OptFunc func(*Auth) error

// WithAuthCookieExpireDuration set auth cookie expiration
func WithAuthCookieExpireDuration(d time.Duration) OptFunc {
	return func(opt *Auth) error {
		if d < 0 {
			return fmt.Errorf("duration should not less than 0, got %v", d)
		}

		opt.jwtTokenExpireDuration = d
		return nil
	}
}

// Auth JWT cookie based token generator and validator.
// Cookie looks like <defaultAuthTokenName>:`{<defaultAuthUserIDCtxKey>: "xxxx"}`
type Auth struct {
	jwt                    *gutils.JWT
	jwtTokenExpireDuration time.Duration
}

// NewAuth create new Auth
func NewAuth(secret []byte, opts ...OptFunc) (a *Auth, err error) {
	var j *gutils.JWT
	if j, err = gutils.NewJWT(
		gutils.WithJWTSignMethod(gutils.SignMethodHS256),
		gutils.WithJWTSecretByte(secret),
	); err != nil {
		return nil, errors.Wrap(err, "try to create Auth got error")
	}

	a = &Auth{
		jwtTokenExpireDuration: defaultJWTTokenExpireDuration,
		jwt:                    j,
	}
	for _, optf := range opts {
		if err = optf(a); err != nil {
			return nil, errors.Wrap(err, "set option")
		}
	}
	return a, nil
}

// GetUserClaims get token from request.ctx then validate and return userid
func (a *Auth) GetUserClaims(ctx context.Context, claims jwt.Claims) (err error) {
	var token string
	if token, err = library.GetGinCtxFromStdCtx(ctx).Cookie(defaultTokenName); err != nil {
		return errors.New("jwt token not found")
	}

	if err = a.jwt.ParseClaims(token, claims); err != nil {
		return errors.Wrap(err, "token invalidate")
	}

	return nil
}

type authCookieOption struct {
	maxAge           int
	path, host       string
	secure, httpOnly bool
}

// CookieOptFunc auth cookie options
type CookieOptFunc func(*authCookieOption) error

// WithCookieMaxAge set auth cookie's maxAge
func WithCookieMaxAge(maxAge int) CookieOptFunc {
	return func(opt *authCookieOption) error {
		if maxAge < 0 {
			return fmt.Errorf("maxAge should not less than 0, got %v", maxAge)
		}

		opt.maxAge = maxAge
		return nil
	}
}

// WithCookiePath set auth cookie's path
func WithCookiePath(path string) CookieOptFunc {
	gutils.Logger.Debug("set auth cookie path", zap.String("path", path))
	return func(opt *authCookieOption) error {
		opt.path = path
		return nil
	}
}

// WithCookieSecure set auth cookie's secure
func WithCookieSecure(secure bool) CookieOptFunc {
	gutils.Logger.Debug("set auth cookie secure", zap.Bool("secure", secure))
	return func(opt *authCookieOption) error {
		opt.secure = secure
		return nil
	}
}

// WithCookieHTTPOnly set auth cookie's HTTPOnly
func WithCookieHTTPOnly(httpOnly bool) CookieOptFunc {
	gutils.Logger.Debug("set auth cookie httpOnly", zap.Bool("httpOnly", httpOnly))
	return func(opt *authCookieOption) error {
		opt.httpOnly = httpOnly
		return nil
	}
}

// WithCookieHost set auth cookie's host
func WithCookieHost(host string) CookieOptFunc {
	gutils.Logger.Debug("set auth cookie host", zap.String("host", host))
	return func(opt *authCookieOption) error {
		opt.host = host
		return nil
	}
}

// SetLoginCookie set jwt token to cookies
func (a *Auth) SetLoginCookie(ctx context.Context, claims jwt.Claims, opts ...CookieOptFunc) (err error) {
	gutils.Logger.Debug("SetLoginCookie")
	ctx2 := library.GetGinCtxFromStdCtx(ctx)

	opt := &authCookieOption{
		maxAge:   int(a.jwtTokenExpireDuration.Seconds()),
		path:     defaultCookiePath,
		secure:   defaultCookieSecure,
		httpOnly: defaultCookieHTTPOnly,
		host:     ctx2.Request.Host,
	}
	if ctx2.Request.URL.Port() != "" {
		opt.host += ":" + ctx2.Request.URL.Port()
	}

	for _, optf := range opts {
		if err = optf(opt); err != nil {
			return errors.Wrap(err, "set option")
		}
	}

	var token string
	if token, err = a.jwt.Sign(claims); err != nil {
		return errors.Wrap(err, "try to generate token got error")
	}

	ctx2.SetCookie(defaultTokenName, token, opt.maxAge, opt.path, opt.host, opt.secure, opt.httpOnly)
	return nil
}
