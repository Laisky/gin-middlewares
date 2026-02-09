package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSetCookie_AppliesOptions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "https://api.example.com/login", nil)

	err := SetCookie(ctx, "sid", "token",
		WithCookieMaxAge(600),
		WithCookiePath("/auth"),
		WithCookieSecure(true),
		WithCookieHTTPOnly(true),
	)
	require.NoError(t, err)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "sid", cookies[0].Name)
	require.Equal(t, "/auth", cookies[0].Path)
	require.Equal(t, "api.example.com", cookies[0].Domain)
	require.Equal(t, 600, cookies[0].MaxAge)
	require.True(t, cookies[0].Secure)
	require.True(t, cookies[0].HttpOnly)
}

func TestSetCookie_WithCookieMaxAgeValidation(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "https://api.example.com/login", nil)

	err := SetCookie(ctx, "sid", "token", WithCookieMaxAge(-1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "maxAge should not less than 0")
	require.Empty(t, w.Header().Values("Set-Cookie"))
}

func TestSetCookie_DefaultDomainStripsPort(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "http://example.com:8080/resource", nil)
	req.Host = "example.com:8080"
	ctx.Request = req

	err := SetCookie(ctx, "sid", "token")
	require.NoError(t, err)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "example.com", cookies[0].Domain)
	require.NotContains(t, w.Header().Get("Set-Cookie"), ":8080")
}

func TestSetCookie_OptionHostStripsPort(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "https://api.example.com/login", nil)

	err := SetCookie(ctx, "sid", "token", WithCookieHost("tenant.example.com:8443"))
	require.NoError(t, err)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "tenant.example.com", cookies[0].Domain)
}

func TestSetCookie_IPv6HostOmitsDomain(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "http://[::1]:8080/resource", nil)
	req.Host = "[::1]:8080"
	ctx.Request = req

	err := SetCookie(ctx, "sid", "token")
	require.NoError(t, err)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Empty(t, cookies[0].Domain)
	require.NotContains(t, w.Header().Get("Set-Cookie"), "Domain=")
}

func TestNormalizeCookieHost(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"empty":               {input: "", expected: ""},
		"plain host":          {input: "example.com", expected: "example.com"},
		"host with port":      {input: "example.com:8080", expected: "example.com"},
		"url value":           {input: "https://example.com:8080/path", expected: "example.com"},
		"ipv4 with port":      {input: "127.0.0.1:8080", expected: "127.0.0.1"},
		"ipv6 with port":      {input: "[::1]:8080", expected: ""},
		"raw ipv6":            {input: "::1", expected: ""},
		"invalid host with :": {input: "example.com:abc", expected: "example.com"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, normalizeCookieHost(tc.input))
		})
	}
}
