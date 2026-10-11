package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestExtractTokenFromAuthHeader(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		header string
		want   string
	}{
		{name: "empty", header: "", want: ""},
		{name: "raw token", header: "abc", want: "abc"},
		{name: "raw token with spaces", header: "  abc  ", want: "abc"},
		{name: "bearer standard", header: "Bearer abc", want: "abc"},
		{name: "bearer with extra spaces", header: "  Bearer   abc  ", want: "abc"},
		{name: "bearer only", header: "Bearer", want: ""},
		{name: "bearer no delimiter", header: "Bearerabc", want: "Bearerabc"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, extractTokenFromAuthHeader(tc.header))
		})
	}
}

func TestGetUserClaims_HeaderVariants(t *testing.T) {
	a, err := NewAuth([]byte(testHS256Secret))
	require.NoError(t, err)

	token, err := a.Sign(&dummyClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1"}})
	require.NoError(t, err)

	cases := []struct {
		name      string
		headerVal string
		wantErr   bool
		wantSub   string
	}{
		{name: "bearer standard", headerVal: "Bearer " + token, wantSub: "user-1"},
		{name: "bearer with extra spaces", headerVal: "  Bearer   " + token + "  ", wantSub: "user-1"},
		{name: "raw token", headerVal: token, wantSub: "user-1"},
		{name: "malformed bearer only", headerVal: "Bearer", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			ctx.Request.Header.Set(authHeaderName, tc.headerVal)

			claims := &dummyClaims{}
			err := a.GetUserClaims(ctx, claims)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantSub, claims.Subject)
		})
	}
}
