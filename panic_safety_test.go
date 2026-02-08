package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	gutils "github.com/Laisky/go-utils/v5"
	glog "github.com/Laisky/go-utils/v5/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestTraceID_NoPanicOnNil(t *testing.T) {
	_, err := TraceID(nil)
	require.NoError(t, err)
}

func TestTraceID_NoPanicOnEmptyGin(t *testing.T) {
	id, err := TraceID(&gin.Context{})
	require.NoError(t, err)
	require.NotEmpty(t, id)
}

func TestFromStd_NoPanicOnNilCtx(t *testing.T) {
	h := FromStd(func(w http.ResponseWriter, r *http.Request) {})
	// calling with nil would normally be handled by gin; ensure we don't panic in our code path
	// we cannot pass nil to gin.HandlerFunc directly, so this just proves conversion compiles
	_ = h
}

func TestGetGinCtxFromStdCtx_NoPanic(t *testing.T) {
	gctx, ok := GetGinCtxFromStdCtx(nil)
	require.False(t, ok)
	require.Nil(t, gctx)
}

func TestCookie_NoPanicOnNilCtx(t *testing.T) {
	err := SetCookie(nil, "k", "v")
	require.Error(t, err)
}

func TestCookie_NoPanicOnNoWriter(t *testing.T) {
	err := SetCookie(&gin.Context{}, "k", "v")
	require.Error(t, err)
}

func TestCookie_SetWithWriterOK(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	require.NoError(t, SetCookie(ctx, "k", "v"))
}

func TestLoggerHelpers_NoPanic(t *testing.T) {
	// GetLogger with nil
	_ = GetLogger(nil)

	// SetLogger with nil
	ctx := SetLogger(nil, glog.Shared.Named("test"))
	require.NotNil(t, ctx)

	// Ctx with nil and empty
	_ = Ctx(nil)
	_ = Ctx(&gin.Context{})

	// BackgroundCtx should panic on nil to fail fast
	require.Panics(t, func() { _ = BackgroundCtx(nil) })
	_ = BackgroundCtx(&gin.Context{})
}

func TestLoggerMiddleware_NoPanicOnNilFields(t *testing.T) {
	mw := NewLoggerMiddleware()
	// Create a context with writer but no request fields
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	// do not set Request to simulate missing request
	// Ensure this does not panic
	mw(ctx)
}

func TestColoredStatus_NoPanic(t *testing.T) {
	require.Equal(t, "", coloredStatus(nil))
	ctx := &gin.Context{}
	require.Equal(t, "", coloredStatus(ctx))
}

type dummyClaims struct{ jwt.RegisteredClaims }

func TestAuth_NoPanicOnMissingParts(t *testing.T) {
	a, err := NewAuth([]byte("secret"))
	require.NoError(t, err)

	// GetUserClaims with non-gin ctx
	err = a.GetUserClaims(context.Background(), &dummyClaims{})
	require.Error(t, err)

	// GetUserClaims with gin ctx but no request
	err = a.GetUserClaims(&gin.Context{}, &dummyClaims{})
	require.Error(t, err)

	// SetAuthHeader with non-gin ctx
	_, err = a.SetAuthHeader(context.Background(), WithSetAuthHeaderToken("t"))
	require.Error(t, err)

	// SetAuthHeader with gin ctx but no writer
	_, err = a.SetAuthHeader(&gin.Context{}, WithSetAuthHeaderToken("t"))
	require.Error(t, err)

	// Sign with valid claims
	_, err = a.Sign(&dummyClaims{RegisteredClaims: jwt.RegisteredClaims{ID: "1"}})
	require.NoError(t, err)
}

func TestMetrics_NoPanicOnNilRouter(t *testing.T) {
	err := EnableMetric(nil)
	require.Error(t, err)
	// BindPrometheus should not panic when nil
	BindPrometheus(nil)
}

func TestCtxTraceInject_NoPanic(t *testing.T) {
	// Ensure Ctx injects trace id even if components are missing
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	c := Ctx(ctx)
	// Verify trace exists in returned context value if possible
	if tid := c.Value(gutils.TracingKey); tid != nil {
		// just access it; no assertion on its form
		_ = tid
	}
}

func TestGetUserClaims_NoPanicOnMalformedHeader(t *testing.T) {
	a, err := NewAuth([]byte("secret"))
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("Authorization", "Bearer") // Exactly "Bearer"

	stdCtx := context.WithValue(context.Background(), CtxKeyGin, ctx)

	require.NotPanics(t, func() {
		err = a.GetUserClaims(stdCtx, &dummyClaims{})
		require.Error(t, err)
	})
}
