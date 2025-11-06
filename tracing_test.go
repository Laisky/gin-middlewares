package middlewares

import (
	"net/http"
	"testing"

	gutils "github.com/Laisky/go-utils/v6"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTraceID(t *testing.T) {
	t.Run("trace id from context", func(t *testing.T) {
		ctx := &gin.Context{}
		expectedID := "1234567890:1234567890:1234:1"
		ctx.Set(gutils.TracingKey, expectedID)

		id, err := TraceID(ctx)
		require.NoError(t, err)
		require.Equal(t, gutils.JaegerTracingID(expectedID), id)
	})

	t.Run("trace id from header", func(t *testing.T) {
		ctx := &gin.Context{
			Request: &http.Request{
				Header: http.Header{},
			},
		}
		expectedID := "1234567890:1234567890:1234:2"
		ctx.Request.Header.Set(gutils.TracingKey, expectedID)

		id, err := TraceID(ctx)
		require.NoError(t, err)
		require.Equal(t, gutils.JaegerTracingID(expectedID), id)
	})

	t.Run("generate new trace id", func(t *testing.T) {
		ctx := &gin.Context{
			Request: &http.Request{
				Header: http.Header{},
			},
		}

		id, err := TraceID(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// Verify the ID was set in context
		storedID := ctx.GetString(gutils.TracingKey)
		require.Equal(t, string(id), storedID)
	})
}
