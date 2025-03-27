package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFromStd(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	ctx.Request = req // Set the request in gin context

	handlerCalled := false
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Verify gin context is embedded
		ginCtx, ok := r.Context().Value(CtxKeyGin).(*gin.Context)
		assert.True(t, ok)
		assert.NotNil(t, ginCtx)
	})

	// Convert and execute handler
	ginHandler := FromStd(mockHandler)
	ginHandler(ctx)

	assert.True(t, handlerCalled, "Handler should have been called")
}

func TestGetGinCtxFromStdCtx(t *testing.T) {
	t.Run("direct gin context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ginCtx, ok := GetGinCtxFromStdCtx(ctx)
		assert.True(t, ok)
		assert.Equal(t, ctx, ginCtx)
	})

	t.Run("context with embedded gin context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)

		ctx := context.WithValue(context.Background(), CtxKeyGin, ginCtx)
		resultGinCtx, ok := GetGinCtxFromStdCtx(ctx)
		assert.True(t, ok)
		assert.Equal(t, ginCtx, resultGinCtx)
	})

	t.Run("invalid context", func(t *testing.T) {
		ctx := context.Background()
		ginCtx, ok := GetGinCtxFromStdCtx(ctx)
		assert.False(t, ok)
		assert.Nil(t, ginCtx)
	})
}
