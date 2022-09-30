package Lplot

import (
	"Lplot/base"
	"Lplot/env"
	"Lplot/middleware"
	"Lplot/server/unix"
	"github.com/gin-gonic/gin"
)

type BootstrapConf struct {
	Pprof bool `yaml:"pprof"`
}

func Bootstraps(router *gin.Engine, conf BootstrapConf) {
	// 环境判断 env GIN_MODE=release/debug
	gin.SetMode(env.RunMode)

	// Global middleware
	router.Use(middleware.Metadata())
	router.Use(middleware.AccessLog())
	router.Use(gin.Recovery())

	// unix socket
	if env.IsDockerPlatform() {
		unix.Start(router)
	}

	// 性能分析工具
	if conf.Pprof {
		base.RegisterProf()
	}

	// 就绪探针
	router.GET("/ready", base.ReadyProbe())
}
