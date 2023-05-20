package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// 获取本机ip
func GetLocalIp() string {
	addrs, _ := net.InterfaceAddrs()
	var ip string = ""
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				if ip != "127.0.0.1" {
					return ip
				}
			}
		}
	}
	return "127.0.0.1"
}

// 过 ingress 的请求clientIp 优先从 "X-Original-Forwarded-For" 中获取
func GetClientIp(ctx *gin.Context) string {
	clientIP := ctx.GetHeader("X-Original-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP != "" {
		return clientIP
	}
	return ctx.ClientIP()
}
