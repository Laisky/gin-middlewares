package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	gutils "github.com/Laisky/go-utils/v6"
	glog "github.com/Laisky/go-utils/v6/log"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSetLogger(t *testing.T) {
	gctx := &gin.Context{}
	logger := glog.Shared.Named("test")

	ctx := SetLogger(gctx, logger)

	getLoggerFromGctx := GetLogger(gctx)
	getLoggerFromCtx := GetLogger(ctx)

	require.Equal(t, logger, getLoggerFromGctx)
	require.Equal(t, logger, getLoggerFromCtx)
}

func TestRequestBasicFieldsRedactsSensitiveQueryParams(t *testing.T) {
	gctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gctx.Request = httptest.NewRequest("GET", "/callback?code=oauth-code&next=%2Fhome&token=bearer-token", nil)
	gctx.Request.RemoteAddr = "127.0.0.1:1234"
	gctx.Request.Host = "example.com"

	urlStr, remote, host := requestBasicFields(gctx)

	require.Equal(t, "/callback?code=%5BREDACTED%5D&next=%2Fhome&token=%5BREDACTED%5D", urlStr)
	require.Equal(t, "127.0.0.1:1234", remote)
	require.Equal(t, "example.com", host)
	require.NotContains(t, urlStr, "oauth-code")
	require.NotContains(t, urlStr, "bearer-token")
}

func TestBackgroundCtxUsesRealBackgroundContext(t *testing.T) {
	t.Parallel()

	type requestCtxKey string

	const key requestCtxKey = "request-only"

	requestBaseCtx, cancel := context.WithCancel(context.WithValue(context.Background(), key, "request-value"))
	defer cancel()

	w := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(w)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/", nil).WithContext(requestBaseCtx)
	logger := glog.Shared.Named("background-ctx")
	SetLogger(gctx, logger)

	ctx := BackgroundCtx(gctx)
	cancel()

	select {
	case <-ctx.Done():
		t.Fatal("BackgroundCtx should not inherit request cancellation")
	default:
	}

	require.Nil(t, ctx.Value(key))
	require.Equal(t, logger, GetLogger(ctx))
	require.Same(t, gctx, ctx.Value(CtxKeyGin))
	require.NotNil(t, ctx.Value(gutils.TracingKey))
	_, ok := GetGinCtxFromStdCtx(ctx)
	require.True(t, ok)
}
