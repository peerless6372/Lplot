package redis

import (
	"errors"
	"github.com/peerless6372/Lplot/klog"
	"github.com/peerless6372/Lplot/utils"
	"time"

	"lzh/gin-gonic/gin"
)

type Pipeliner interface {
	Exec(ctx *gin.Context) ([]interface{}, error)
	Put(ctx *gin.Context, cmd string, args ...interface{}) error
}

type commands struct {
	cmd   string
	args  []interface{}
	reply interface{}
	err   error
}

type Pipeline struct {
	cmds  []commands
	err   error
	redis *Redis
}

func (r *Redis) Pipeline() Pipeliner {
	return &Pipeline{
		redis: r,
	}
}

func (p *Pipeline) Put(ctx *gin.Context, cmd string, args ...interface{}) error {
	if len(args) < 1 {
		return errors.New("no key found in args")
	}
	c := commands{
		cmd:  cmd,
		args: args,
	}
	p.cmds = append(p.cmds, c)
	return nil
}

func (p *Pipeline) Exec(ctx *gin.Context) (res []interface{}, err error) {
	start := time.Now()

	addr, conn, err := p.redis.choosePool(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	for i := range p.cmds {
		err = conn.Send(p.cmds[i].cmd, p.cmds[i].args...)
	}

	err = conn.Flush()

	var msg string
	var ralCode int
	if err == nil {
		ralCode = 0
		for i := range p.cmds {
			var reply interface{}
			reply, err = conn.Receive()
			res = append(res, reply)
			p.cmds[i].reply, p.cmds[i].err = reply, err
		}

		msg = "pipeline exec succ"
	} else {
		ralCode = -1
		p.err = err
		msg = "pipeline exec error: " + err.Error()
	}

	end := time.Now()
	fields := []klog.Field{
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("prot", "redis"),
		klog.String("remoteAddr", addr),
		klog.String("service", p.redis.Service),
		klog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		klog.Float64("cost", utils.GetRequestCost(start, end)),
		klog.Int("ralCode", ralCode),
	}

	klog.InfoLogger(ctx, msg, fields...)

	return res, err
}
