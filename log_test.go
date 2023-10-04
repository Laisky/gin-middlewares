package middlewares

import (
	"testing"

	glog "github.com/Laisky/go-utils/v4/log"
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
