package klog

import (
	"github.com/peerless6372/Lplot/env"
	"go.uber.org/zap"
	"lzh/gin-gonic/gin"
)

func GetZapLogger() (l *zap.Logger) {
	if ZapLogger == nil {
		ZapLogger = newLogger().WithOptions(zap.AddCallerSkip(1))
	}
	return ZapLogger
}

func zapLogger(ctx *gin.Context) *zap.Logger {
	m := GetZapLogger()
	//m = m.WithOptions(zap.AddCallerSkip(1))
	if ctx == nil {
		return m
	}
	return m.With(
		zap.String("requestId", GetRequestID(ctx)),
		zap.String("module", env.GetAppName()),
		zap.String("localIp", env.LocalIP),
		zap.String("uri", ctx.GetString(ContextKeyUri)),
	)
}

func DebugLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Debug(msg, fields...)
}
func InfoLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Info(msg, fields...)
}

func WarnLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Warn(msg, fields...)
}

func ErrorLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Error(msg, fields...)
}

func PanicLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Panic(msg, fields...)
}

func FatalLogger(ctx *gin.Context, msg string, fields ...zap.Field) {
	if NoLog(ctx) {
		return
	}
	zapLogger(ctx).Fatal(msg, fields...)
}
