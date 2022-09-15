// Package middlewares useful middlewares
package middlewares

import (
	"context"
	"fmt"
	"strings"

	gjwt "github.com/Laisky/go-utils/v2/jwt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

const (
	authHeaderName   = "Authorization"
	authHeaderPrefix = "Bearer"
	authHeaderLayout = authHeaderPrefix + " %s"
)

// AuthOptFunc auth option
type AuthOptFunc func(*Auth) error

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
	jwt gjwt.JWT
}

// NewAuth create new Auth
func NewAuth(secret []byte, opts ...AuthOptFunc) (a *Auth, err error) {
	a = &Auth{}
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
	token := GetGinCtxFromStdCtx(ctx).GetHeader(authHeaderName)
	if strings.HasPrefix(token, authHeaderPrefix) { // remove "Bearer "
		token = token[len(authHeaderPrefix)+1:]
	}

	if err = a.jwt.ParseClaims(token, claims); err != nil {
		return errors.Wrap(err, "token invalidate")
	}

	return nil
}

type setAuthHeaderOption struct {
	claim jwt.Claims
	token string
}

func (o *setAuthHeaderOption) applyOpts(opts ...SetAuthHeaderOption) (*setAuthHeaderOption, error) {
	for _, f := range opts {
		if err := f(o); err != nil {
			return nil, err
		}
	}

	return o, nil
}

type SetAuthHeaderOption func(*setAuthHeaderOption) error

func WithSetAuthHeaderClaim(claim jwt.Claims) SetAuthHeaderOption {
	return func(o *setAuthHeaderOption) error {
		if claim == nil {
			return errors.Errorf("claim is empty")
		}

		if o.token != "" {
			return errors.Errorf("claim and token should not be set at the same time")
		}

		o.claim = claim
		return nil
	}
}

func WithSetAuthHeaderToken(token string) SetAuthHeaderOption {
	return func(o *setAuthHeaderOption) error {
		if token == "" {
			return errors.Errorf("token is empty")
		}

		if o.claim != nil {
			return errors.Errorf("claim and token should not be set at the same time")
		}

		o.token = token
		return nil
	}
}

// SetAuthHeader set jwt token to cookies
func (a *Auth) SetAuthHeader(ctx context.Context, optfs ...SetAuthHeaderOption) (string, error) {
	opt, err := new(setAuthHeaderOption).applyOpts(optfs...)
	if err != nil {
		return "", err
	}

	if opt.claim == nil && opt.token == "" {
		return "", errors.New("claim or token should be set")
	}

	if opt.token == "" {
		var err error
		opt.token, err = a.Sign(opt.claim)
		if err != nil {
			return "", err
		}
	}

	ginCtx := GetGinCtxFromStdCtx(ctx)
	ginCtx.Header(authHeaderName, fmt.Sprintf(authHeaderLayout, opt.token))
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
