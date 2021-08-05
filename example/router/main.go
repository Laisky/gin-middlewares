package main

import (
	"time"

	gm "github.com/Laisky/gin-middlewares"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	if err := gm.EnableMetric(engine,
		gm.WithMetricAddr("127.0.0.1:8080"),
		gm.WithPprofPath("/pprof"),
		gm.WithMetricGraceWait(1*time.Second),
	); err != nil {
		gutils.Logger.Panic("enable metrics", zap.Error(err))
	}

	gutils.Logger.Info("start server at 127.0.0.1:8080")
	gutils.Logger.Info("exit", zap.Error(engine.Run("127.0.0.1:8080")))
}
