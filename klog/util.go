package klog

import (
	"github.com/peerless6372/Lplot/utils/metadata"
	"lzh/gin-gonic/gin"
	"strconv"
	"time"
)

// util key
const (
	ContextKeyRequestID = "requestId"
	ContextKeyLogID     = "logID"
	ContextKeyNoLog     = "_no_log"
	ContextKeyUri       = "_uri"
)

// header key
const (
	TraceHeaderKey      = "Uber-Trace-Id"
	LogIDHeaderKey      = "X_BD_LOGID"
	LogIDHeaderKeyLower = "x_bd_logid"
)

func GetRequestID(ctx *gin.Context) string {
	if ctx == nil {
		return genRequestId()
	}

	// 从ctx中获取
	if r := ctx.GetString(ContextKeyRequestID); r != "" {
		return r
	}

	// 优先从header中获取
	var requestId string
	if ctx.Request != nil && ctx.Request.Header != nil {
		requestId = ctx.Request.Header.Get(TraceHeaderKey)
	}

	// 新生成
	if requestId == "" {
		requestId = genRequestId()
	}

	ctx.Set(ContextKeyRequestID, requestId)
	return requestId
}

func genRequestId() (requestId string) {
	// 随机生成 todo: 随机生成的格式是否要统一成trace的格式
	usec := uint64(time.Now().UnixNano())
	requestId = strconv.FormatUint(usec&0x7FFFFFFF|0x80000000, 10)
	return requestId
}

// 用户自定义Notice
func AddNotice(ctx *gin.Context, key string, val interface{}) {
	if meta, ok := metadata.CtxFromGinContext(ctx); ok {
		if n := metadata.Value(meta, metadata.Notice); n != nil {
			if _, ok = n.(map[string]interface{}); ok {
				notices := n.(map[string]interface{})
				notices[key] = val
			}
		}
	}
}

// 获得所有用户自定义的Notice
func GetCustomerKeyValue(ctx *gin.Context) map[string]interface{} {
	meta, ok := metadata.CtxFromGinContext(ctx)
	if !ok {
		return nil
	}

	n := metadata.Value(meta, metadata.Notice)
	if n == nil {
		return nil
	}
	if notices, ok := n.(map[string]interface{}); ok {
		return notices
	}

	return nil
}

// server.log 中打印出用户自定义Notice
func PrintNotice(ctx *gin.Context) {
	notices := GetCustomerKeyValue(ctx)

	var fields []interface{}
	for k, v := range notices {
		fields = append(fields, k, v)
	}
	sugaredLogger(ctx).With(fields...).Info("notice")
}

func SetNoLogFlag(ctx *gin.Context) {
	ctx.Set(ContextKeyNoLog, true)
}

func NoLog(ctx *gin.Context) bool {
	if ctx == nil {
		return false
	}
	flag, ok := ctx.Get(ContextKeyNoLog)
	if ok && flag == true {
		return true
	}
	return false
}
