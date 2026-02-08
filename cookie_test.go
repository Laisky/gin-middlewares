package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSetCookie(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "http://example.com/", nil)
		ctx.Request.Host = "example.com"

		err := SetCookie(ctx, "test_cookie", "test_value")
		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		cookie := cookies[0]
		require.Equal(t, "test_cookie", cookie.Name)
		require.Equal(t, "test_value", cookie.Value)
		require.Equal(t, "/", cookie.Path)
		require.Equal(t, "example.com", cookie.Domain)
		require.True(t, cookie.HttpOnly, "HttpOnly should be true by default")
		require.False(t, cookie.Secure, "Secure should be false by default")
	})

	t.Run("custom options", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/", nil)

		err := SetCookie(ctx, "test_cookie", "test_value",
			WithCookieHTTPOnly(false),
			WithCookieSecure(true),
			WithCookieMaxAge(3600),
			WithCookiePath("/api"),
			WithCookieHost("api.example.com"),
		)
		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		cookie := cookies[0]
		require.Equal(t, "test_cookie", cookie.Name)
		require.Equal(t, "test_value", cookie.Value)
		require.Equal(t, "/api", cookie.Path)
		require.Equal(t, "api.example.com", cookie.Domain)
		require.Equal(t, 3600, cookie.MaxAge)
		require.False(t, cookie.HttpOnly, "HttpOnly should be false as requested")
		require.True(t, cookie.Secure, "Secure should be true as requested")
	})

	t.Run("domain with port", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "http://example.com:8080/", nil)
		ctx.Request.Host = "example.com:8080"

		err := SetCookie(ctx, "test_cookie", "test_value")
		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		cookie := cookies[0]
		require.Equal(t, "example.com", cookie.Domain, "Domain should not include port")
	})
}
