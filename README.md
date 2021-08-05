# gin-middlewares

Add Promethues metrics and pprof to your Gin application in one line.

separated from `go-utils/gin-middlewares`

## Metrics Server

You can find example codes in `./example`.

### Standalone

If you do not has a gin server, you can use `NewHTTPMetricSrv` to start an new gin HTTP server with metrics.

```go
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

```

### Route

If you already has a gin server, you can add metrics route in your gin server.

```go
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
		gm.WithPprofPath("/pprof"),
		gm.WithMetricGraceWait(1*time.Second),
	); err != nil {
		gutils.Logger.Panic("enable metrics", zap.Error(err))
	}

	gutils.Logger.Info("start server at 127.0.0.1:8080")
	gutils.Logger.Info("exit", zap.Error(engine.Run("127.0.0.1:8080")))
}
```
