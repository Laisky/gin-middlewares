package middlewares

import (
	"net/http"

	glog "github.com/Laisky/go-utils/v2/log"
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
		glog.Shared.Panic("try to init gin auth got error", zap.Error(err))
	}

	ctx := &gin.Context{}
	uc := &UserClaims{}
	if err := auth.GetUserClaims(ctx, uc); err != nil {
		glog.Shared.Warn("user invalidate", zap.Error(err))
	} else {
		glog.Shared.Info("user validate", zap.String("uid", uc.Subject))
	}

	if _, err = auth.SetLoginCookie(ctx, WithAuthClaims(uc)); err != nil {
		glog.Shared.Error("try to set cookie got error", zap.Error(err))
	}

	Server := gin.New()
	Server.Handle("ANY", "/authorized/", FromStd(DemoHandle))
}

func DemoHandle(w http.ResponseWriter, r *http.Request) {
	// middlewares
	if _, err := w.Write([]byte("hello")); err != nil {
		glog.Shared.Error("http write", zap.Error(err))
	}
}
