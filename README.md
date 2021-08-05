# gin-middlewares

Add Promethues metrics and pprof to your Gin application in one line.

separated from `go-utils/gin-middlewares`

## Metrics Server

You can find example codes in `./example/metrics/`.

### Standalone

If you do not has a gin server, you can use `NewHTTPMetricSrv` to start an new gin HTTP server with metrics.

```go
package main

import (
	"context"
	"log"
	"time"

	gm "github.com/Laisky/gin-middlewares/metrics"
)

func main() {
	ctx := context.Background()
	srv, err := gm.NewHTTPMetricSrv(ctx,
		gm.WithMetricAddr("127.0.0.1:8080"),
		gm.WithPprofPath("/pprof"),
		gm.WithMetricGraceWait(1*time.Second),
	)
	if err != nil {
		log.Panic("new metric server", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", srv.ListenAndServe())
}


```

### Route

If you already has a gin server, you can add metrics route in your gin server.

```go
package main

import (
	"log"
	"time"

	gm "github.com/Laisky/gin-middlewares/metrics"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	if err := gm.EnableMetric(engine,
		gm.WithMetricAddr("127.0.0.1:8080"),
		gm.WithPprofPath("/pprof"),
		gm.WithMetricGraceWait(1*time.Second),
	); err != nil {
		log.Panic("enable metrics", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", engine.Run("127.0.0.1:8080"))
}

```
