// Package middlewares useful middlewares
package middlewares

import (
	"context"
	"fmt"
	"time"

	gjwt "github.com/Laisky/go-utils/v2/jwt"
	glog "github.com/Laisky/go-utils/v2/log"
	"github.com/Laisky/zap"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

const (
	// defaultAuthTokenName jwt token cookie name
	defaultAuthTokenName = "token"
	// defaultAuthUserIDCtxKey key of user ID in jwt token
	// defaultAuthUserIDCtxKey           = "auth_uid"
	defaultAuthJWTTokenExpireDuration = 7 * 24 * time.Hour

	defaultAuthCookiePath     = "/"
	defaultAuthCookieSecure   = false
	defaultAuthCookieHTTPOnly = false
)

// AuthOptFunc auth option
type AuthOptFunc func(*Auth) error

// WithAuthCookieExpireDuration set auth cookie expiration
func WithAuthCookieExpireDuration(d time.Duration) AuthOptFunc {
	return func(opt *Auth) error {
		if d < 0 {
			return fmt.Errorf("duration should not less than 0, got %v", d)
		}

		opt.jwtTokenExpireDuration = d
		return nil
	}
}

// WithAuthJWT set jwt lib
func WithAuthJWT(jwt gjwt.JWT) AuthOptFunc {
	return func(opt *Auth) error {
		opt.jwt = jwt
		return nil
	}
}

// Auth JWT cookie based token generator and validator.
// Cookie looks like <defaultAuthTokenName>:`{<defaultAuthUserIDCtxKey>: "xxxx"}`
type Auth struct {
	jwt                    gjwt.JWT
	jwtTokenExpireDuration time.Duration
}

// NewAuth create new Auth
func NewAuth(secret []byte, opts ...AuthOptFunc) (a *Auth, err error) {
	a = &Auth{
		jwtTokenExpireDuration: defaultAuthJWTTokenExpireDuration,
	}
	for _, optf := range opts {
		if err = optf(a); err != nil {
			return nil, errors.Wrap(err, "set option")
		}
	}

	// set default jwt lib
	if a.jwt == nil {
		if a.jwt, err = gjwt.New(
			gjwt.WithSignMethod(gjwt.SignMethodHS256),
			gjwt.WithSecretByte(secret),
		); err != nil {
			return nil, errors.Wrap(err, "try to create Auth got error")
		}
	}

	return a, nil
}

// GetUserClaims get token from request.ctx then validate and return userid
func (a *Auth) GetUserClaims(ctx context.Context, claims jwt.Claims) (err error) {
	var token string
	if token, err = GetGinCtxFromStdCtx(ctx).Cookie(defaultAuthTokenName); err != nil {
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
	claim            jwt.Claims
	token            string
}

// AuthCookieOptFunc auth cookie options
type AuthCookieOptFunc func(*authCookieOption) error

// WithAuthCookieMaxAge set auth cookie's maxAge
func WithAuthCookieMaxAge(maxAge int) AuthCookieOptFunc {
	return func(opt *authCookieOption) error {
		if maxAge < 0 {
			return fmt.Errorf("maxAge should not less than 0, got %v", maxAge)
		}

		opt.maxAge = maxAge
		return nil
	}
}

// WithAuthCookiePath set auth cookie's path
func WithAuthCookiePath(path string) AuthCookieOptFunc {
	glog.Shared.Debug("set auth cookie path", zap.String("path", path))
	return func(opt *authCookieOption) error {
		opt.path = path
		return nil
	}
}

// WithAuthCookieSecure set auth cookie's secure
func WithAuthCookieSecure(secure bool) AuthCookieOptFunc {
	glog.Shared.Debug("set auth cookie secure", zap.Bool("secure", secure))
	return func(opt *authCookieOption) error {
		opt.secure = secure
		return nil
	}
}

// WithAuthCookieHTTPOnly set auth cookie's HTTPOnly
func WithAuthCookieHTTPOnly(httpOnly bool) AuthCookieOptFunc {
	glog.Shared.Debug("set auth cookie httpOnly", zap.Bool("httpOnly", httpOnly))
	return func(opt *authCookieOption) error {
		opt.httpOnly = httpOnly
		return nil
	}
}

// WithAuthCookieHost set auth cookie's host
func WithAuthCookieHost(host string) AuthCookieOptFunc {
	glog.Shared.Debug("set auth cookie host", zap.String("host", host))
	return func(opt *authCookieOption) error {
		opt.host = host
		return nil
	}
}

// WithAuthClaims set claims that will used to sign jwt token
func WithAuthClaims(claims jwt.Claims) AuthCookieOptFunc {
	return func(opt *authCookieOption) error {
		opt.claim = claims
		return nil
	}
}

// WIthAuthToken set jwt token to response
func WithAuthToken(token string) AuthCookieOptFunc {
	return func(opt *authCookieOption) error {
		opt.token = token
		return nil
	}
}

// SetLoginCookie set jwt token to cookies
func (a *Auth) SetLoginCookie(ctx context.Context,
	opts ...AuthCookieOptFunc) (token string, err error) {
	glog.Shared.Debug("SetLoginCookie")
	ctx2 := GetGinCtxFromStdCtx(ctx)

	opt := &authCookieOption{
		maxAge:   int(a.jwtTokenExpireDuration.Seconds()),
		path:     defaultAuthCookiePath,
		secure:   defaultAuthCookieSecure,
		httpOnly: defaultAuthCookieHTTPOnly,
		host:     ctx2.Request.Host,
	}
	if ctx2.Request.URL.Port() != "" {
		opt.host += ":" + ctx2.Request.URL.Port()
	}

	for _, optf := range opts {
		if err = optf(opt); err != nil {
			return "", errors.Wrap(err, "set option")
		}
	}

	if opt.claim == nil && opt.token == "" {
		return "", errors.New("claim or token should be set")
	} else if opt.claim != nil && opt.token != "" {
		return "", errors.New("claim and token should not be set at the same time")
	}

	if opt.claim != nil {
		if opt.token, err = a.Sign(opt.claim); err != nil {
			return "", err
		}
	}

	ctx2.SetCookie(defaultAuthTokenName, opt.token, opt.maxAge, opt.path, opt.host, opt.secure, opt.httpOnly)
	return opt.token, nil
}

// Sign sign jwt token
func (a *Auth) Sign(claim jwt.Claims) (string, error) {
	token, err := a.jwt.Sign(claim)
	if err != nil {
		return "", errors.Wrap(err, "try to generate token got error")
	}

	return token, nil
}
