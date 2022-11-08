package middlewares

import (
	log "github.com/Laisky/go-utils/v3/log"
	"github.com/Laisky/zap"
)

var Logger log.Logger

func init() {
	var err error
	Logger, err = log.NewConsoleWithName("gin-mw", log.LevelInfo)
	if err != nil {
		log.Shared.Panic("new log", zap.Error(err))
	}
}
