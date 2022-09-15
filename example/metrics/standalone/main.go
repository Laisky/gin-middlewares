// Package main run metric server
package main

import (
	"context"
	"log"
	"time"

	ginMw "github.com/Laisky/gin-middlewares/v3"
)

func main() {
	ctx := context.Background()
	srv, err := ginMw.NewHTTPMetricSrv(ctx,
		ginMw.WithMetricAddr("127.0.0.1:8080"),
		ginMw.WithPprofPath("/pprof"),
		ginMw.WithMetricGraceWait(1*time.Second),
	)
	if err != nil {
		log.Panic("new metric server", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", srv.ListenAndServe())
}
