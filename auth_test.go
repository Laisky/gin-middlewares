package middlewares

import (
	"net/http"

	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type UserClaims struct {
	jwt.StandardClaims
}

func ExampleAuth() {
	auth, err := NewAuth([]byte("f32lifj2f32fj"))
	if err != nil {
		Logger.Panic("try to init gin auth got error", zap.Error(err))
	}

	ctx := &gin.Context{}
	uc := &UserClaims{}
	if err := auth.GetUserClaims(ctx, uc); err != nil {
		Logger.Warn("user invalidate", zap.Error(err))
	} else {
		Logger.Info("user validate", zap.String("uid", uc.Subject))
	}

	if _, err = auth.SetAuthHeader(ctx, WithSetAuthHeaderClaim(uc)); err != nil {
		Logger.Error("try to set cookie got error", zap.Error(err))
	}

	Server := gin.New()
	Server.Handle("ANY", "/authorized/", FromStd(DemoHandle))
}

func DemoHandle(w http.ResponseWriter, r *http.Request) {
	// middlewares
	if _, err := w.Write([]byte("hello")); err != nil {
		Logger.Error("http write", zap.Error(err))
	}
}
