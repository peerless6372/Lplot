package redis

import (
	"github.com/peerless6372/Lplot/klog"
	"github.com/peerless6372/Lplot/utils"
	"time"

	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
)

func (r *Redis) Lua(ctx *gin.Context, script string, keyCount int, keysAndArgs ...interface{}) (interface{}, error) {
	start := time.Now()

	lua := redigo.NewScript(keyCount, script)

	addr, conn, err := r.choosePool(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	reply, err := lua.Do(conn, keysAndArgs...)

	ralCode := 0
	msg := "pipeline exec succ"
	if err != nil {
		ralCode = -1
		msg = "pipeline exec error: " + err.Error()
	}
	end := time.Now()

	fields := []klog.Field{
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("prot", "redis"),
		klog.String("remoteAddr", addr),
		klog.String("service", r.Service),
		klog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		klog.Float64("cost", utils.GetRequestCost(start, end)),
		klog.Int("ralCode", ralCode),
	}

	klog.InfoLogger(ctx, msg, fields...)

	return reply, err
}
