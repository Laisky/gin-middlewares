package main

import (
	"context"
	"time"

	gm "github.com/Laisky/gin-middlewares"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
)

func main() {
	ctx := context.Background()
	srv, err := gm.NewHTTPMetricSrv(ctx,
		gm.WithMetricAddr("127.0.0.1:8080"),
		gm.WithPprofPath("/pprof"),
		gm.WithMetricGraceWait(1*time.Second),
	)
	if err != nil {
		gutils.Logger.Panic("new metric server", zap.Error(err))
	}

	gutils.Logger.Info("start server at 127.0.0.1:8080")
	gutils.Logger.Info("exit", zap.Error(srv.ListenAndServe()))
}
