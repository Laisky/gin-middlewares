package main

import (
	"context"
	"log"
	"time"

	gm "github.com/Laisky/gin-middlewares/metrics"
)

func main() {
	ctx := context.Background()
	srv, err := gm.NewHTTPSrv(ctx,
		gm.WithAddr("127.0.0.1:8080"),
		gm.WithPprofPath("/pprof"),
		gm.WithGraceWait(1*time.Second),
	)
	if err != nil {
		log.Panic("new metric server", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", srv.ListenAndServe())
}
