package main

import (
	"log"
	"time"

	gm "github.com/Laisky/gin-middlewares/metrics"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	if err := gm.Enable(engine,
		gm.WithPprofPath("/pprof"),
		gm.WithGraceWait(1*time.Second),
	); err != nil {
		log.Panic("enable metrics", err)
	}

	log.Println("start server at 127.0.0.1:8080")
	log.Println("exit", engine.Run("127.0.0.1:8080"))
}
