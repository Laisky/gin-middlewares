package auth

import (
	"net/http"

	"github.com/Laisky/gin-middlewares/library"
	"github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

type UserClaims struct {
	jwt.StandardClaims
}

func ExampleAuth() {
	auth, err := NewAuth([]byte("f32lifj2f32fj"))
	if err != nil {
		utils.Logger.Panic("try to init gin auth got error", zap.Error(err))
	}

	ctx := &gin.Context{}
	uc := &UserClaims{}
	if err := auth.GetUserClaims(ctx, uc); err != nil {
		utils.Logger.Warn("user invalidate", zap.Error(err))
	} else {
		utils.Logger.Info("user validate", zap.String("uid", uc.Subject))
	}

	if err = auth.SetLoginCookie(ctx, uc); err != nil {
		utils.Logger.Error("try to set cookie got error", zap.Error(err))
	}

	Server := gin.New()
	Server.Handle("ANY", "/authorized/", library.FromStd(DemoHandle))
}

func DemoHandle(w http.ResponseWriter, r *http.Request) {
	// middlewares
	if _, err := w.Write([]byte("hello")); err != nil {
		utils.Logger.Error("http write", zap.Error(err))
	}
}
