package http

import (
	"github.com/peerless6372/gin"
)

func Start(engine *gin.Engine, conf ServerConfig) error {
	return engine.Run(conf.Address)
}
