package http

import (
	"github.com/gin-gonic/gin"
)

func Start(engine *gin.Engine, conf ServerConfig) error {
	return engine.Run(conf.Address)
}
