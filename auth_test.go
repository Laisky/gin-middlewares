package middlewares

import (
	"net/http"

	"github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
)

type User struct{}

func (u *User) GetPayload() map[string]interface{} {
	return map[string]interface{}{"a": "b"}
}

func (u *User) GetID() string {
	return "123"
}

func ExampleAuth() {
	auth, err := NewAuth([]byte("f32lifj2f32fj"))
	if err != nil {
		utils.Logger.Panic("try to init gin auth got error", zap.Error(err))
	}

	ctx := &gin.Context{}
	uid, err := auth.ValidateAndGetUID(ctx)
	if err != nil {
		utils.Logger.Warn("user invalidate", zap.Error(err))
	} else {
		utils.Logger.Info("user validate", zap.String("uid", uid))
	}

	user := &User{}
	if err = auth.SetLoginCookie(ctx, user); err != nil {
		utils.Logger.Error("try to set cookie got error", zap.Error(err))
	}

	Server := gin.New()
	Server.Handle("ANY", "/authorized/", FromStd(DemoHandle))
}

func DemoHandle(w http.ResponseWriter, r *http.Request) {
	// middlewares
	if _, err := w.Write([]byte("hello")); err != nil {
		utils.Logger.Error("http write", zap.Error(err))
	}
}
