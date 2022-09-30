package http

import (
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func Start(engine *gin.Engine, conf ServerConfig) error {
	// todo: 后续根据环境区分，正式环境不允许用户指定端口
	appServer := endless.NewServer(conf.Address, engine)

	// 超时时间 (如果设置的太小，可能导致接口响应时间超过该值，进而导致504错误)
	if conf.ReadTimeout > 0 {
		appServer.ReadTimeout = conf.ReadTimeout
	}

	if conf.WriteTimeout > 0 {
		appServer.WriteTimeout = conf.WriteTimeout
	}

	// 监听http端口
	if err := appServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
