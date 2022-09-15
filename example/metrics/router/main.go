// Package main run metric server as gin's router
package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	ginMw "github.com/Laisky/gin-middlewares/v3"
)

func main() {
	engine := gin.Default()

	if err := ginMw.EnableMetric(engine,
		ginMw.WithPprofPath("/pprof"),
		ginMw.WithMetricGraceWait(1*time.Second),
	); err != nil {
		log.Panic("enable metrics", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", engine.Run("127.0.0.1:8080"))
}
