package middlewares

import (
	log "github.com/Laisky/go-utils/v5/log"
	"github.com/Laisky/zap"
)

var Logger log.Logger

func init() {
	var err error
	Logger, err = log.NewConsoleWithName("gin-mw", log.LevelInfo)
	if err != nil {
		// Do not panic in fundamental packages; fallback to shared logger
		log.Shared.Error("init console logger failed, fallback to shared", zap.Error(err))
		Logger = log.Shared.Named("gin-mw")
	}
}
