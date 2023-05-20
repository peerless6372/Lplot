package base

import "github.com/gin-gonic/gin"

type Probe struct {
	health *gin.HandlerFunc
	ready  *gin.HandlerFunc
}

var p Probe

func RegHealthProbe(h gin.HandlerFunc) {
	p.health = &h
}

func RegReadyProbe(h gin.HandlerFunc) {
	p.ready = &h
}

func HealthProbe() gin.HandlerFunc {
	if p.health == nil {
		return func(c *gin.Context) {
			c.String(200, "succ")
		}
	}
	return *p.health
}

func ReadyProbe() gin.HandlerFunc {
	if p.ready == nil {
		return func(c *gin.Context) {
			c.String(200, "succ")
		}
	}
	return *p.ready
}
