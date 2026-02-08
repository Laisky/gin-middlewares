package middlewares

import (
	"context"
	stderrors "errors"
	"net/http/httptest"
	"testing"

	gjwt "github.com/Laisky/go-utils/v6/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

type mockJWT struct {
	signToken string
	signErr   error
	parseErr  error
}

func (m *mockJWT) Sign(claims jwt.Claims, opts ...gjwt.DivideOption) (string, error) {
	if m.signErr != nil {
		return "", m.signErr
	}
	if m.signToken != "" {
		return m.signToken, nil
	}
	return "mock-token", nil
}

func (m *mockJWT) SignByHS256(claims jwt.Claims, opts ...gjwt.DivideOption) (string, error) {
	return m.Sign(claims, opts...)
}

func (m *mockJWT) SignByES256(claims jwt.Claims, opts ...gjwt.DivideOption) (string, error) {
	return m.Sign(claims, opts...)
}

func (m *mockJWT) ParseClaims(token string, claimsPtr jwt.Claims, opts ...gjwt.DivideOption) error {
	return m.parseErr
}

func (m *mockJWT) ParseClaimsByHS256(token string, claimsPtr jwt.Claims, opts ...gjwt.DivideOption) error {
	return m.ParseClaims(token, claimsPtr, opts...)
}

func (m *mockJWT) ParseClaimsByES256(token string, claimsPtr jwt.Claims, opts ...gjwt.DivideOption) error {
	return m.ParseClaims(token, claimsPtr, opts...)
}

func (m *mockJWT) ParseClaimsByRS256(token string, claimsPtr jwt.Claims, opts ...gjwt.DivideOption) error {
	return m.ParseClaims(token, claimsPtr, opts...)
}

func TestNewAuth_AllBranches(t *testing.T) {
	t.Parallel()

	_, err := NewAuth([]byte("secret"), func(*Auth) error {
		return stderrors.New("opt boom")
	})
	require.ErrorContains(t, err, "set option")

	_, err = NewAuth(nil)
	require.ErrorContains(t, err, "try to create Auth got error")

	mjwt := &mockJWT{}
	a, err := NewAuth([]byte("secret"), WithAuthJWT(mjwt))
	require.NoError(t, err)
	require.Same(t, mjwt, a.jwt)
}

func TestWithSetAuthHeaderClaim_AllBranches(t *testing.T) {
	t.Parallel()

	err := WithSetAuthHeaderClaim(nil)(&setAuthHeaderOption{})
	require.ErrorContains(t, err, "claim is empty")

	err = WithSetAuthHeaderClaim(&dummyClaims{})(&setAuthHeaderOption{token: "existing-token"})
	require.ErrorContains(t, err, "claim and token should not be set at the same time")

	opt := &setAuthHeaderOption{}
	claim := &dummyClaims{}
	require.NoError(t, WithSetAuthHeaderClaim(claim)(opt))
	require.Same(t, claim, opt.claim)
}

func TestWithSetAuthHeaderToken_AllBranches(t *testing.T) {
	t.Parallel()

	err := WithSetAuthHeaderToken("")(&setAuthHeaderOption{})
	require.ErrorContains(t, err, "token is empty")

	err = WithSetAuthHeaderToken("t")(&setAuthHeaderOption{claim: &dummyClaims{}})
	require.ErrorContains(t, err, "claim and token should not be set at the same time")

	opt := &setAuthHeaderOption{}
	require.NoError(t, WithSetAuthHeaderToken("new-token")(opt))
	require.Equal(t, "new-token", opt.token)
}

func TestSetAuthHeader_AllBranches(t *testing.T) {
	t.Parallel()

	a := &Auth{jwt: &mockJWT{}}

	_, err := a.SetAuthHeader(context.Background(), WithSetAuthHeaderToken(""))
	require.ErrorContains(t, err, "token is empty")

	_, err = a.SetAuthHeader(context.Background())
	require.ErrorContains(t, err, "claim or token should be set")

	aSignErr := &Auth{jwt: &mockJWT{signErr: stderrors.New("sign fail")}}
	_, err = aSignErr.SetAuthHeader(context.Background(), WithSetAuthHeaderClaim(&dummyClaims{}))
	require.ErrorContains(t, err, "try to generate token got error")

	_, err = a.SetAuthHeader(context.Background(), WithSetAuthHeaderToken("plain-token"))
	require.ErrorContains(t, err, "gin context not found in ctx")

	_, err = a.SetAuthHeader(&gin.Context{}, WithSetAuthHeaderToken("plain-token"))
	require.ErrorContains(t, err, "gin writer is nil")

	w := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(w)
	token, err := a.SetAuthHeader(gctx, WithSetAuthHeaderToken("plain-token"))
	require.NoError(t, err)
	require.Equal(t, "plain-token", token)
	require.Equal(t, "Bearer plain-token", w.Header().Get(authHeaderName))

	aWithSign := &Auth{jwt: &mockJWT{signToken: "signed-token"}}
	w2 := httptest.NewRecorder()
	gctx2, _ := gin.CreateTestContext(w2)
	token, err = aWithSign.SetAuthHeader(gctx2, WithSetAuthHeaderClaim(&dummyClaims{}))
	require.NoError(t, err)
	require.Equal(t, "signed-token", token)
	require.Equal(t, "Bearer signed-token", w2.Header().Get(authHeaderName))
}

func TestSign_AllBranches(t *testing.T) {
	t.Parallel()

	a := &Auth{jwt: &mockJWT{signToken: "ok-token"}}
	token, err := a.Sign(&dummyClaims{})
	require.NoError(t, err)
	require.Equal(t, "ok-token", token)

	aErr := &Auth{jwt: &mockJWT{signErr: stderrors.New("boom")}}
	_, err = aErr.Sign(&dummyClaims{})
	require.ErrorContains(t, err, "try to generate token got error")
}
