package redis

import (
	"Lplot/env"
	"Lplot/klog"
	"Lplot/utils"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
)

// 日志打印Do args部分支持的最大长度
const logForRedisValue = 50
const prefix = "@@redis."

type RedisConf struct {
	Service         string        `yaml:"service"`
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	MaxIdle         int           `yaml:"maxIdle"`
	MaxActive       int           `yaml:"maxActive"`
	IdleTimeout     time.Duration `yaml:"idleTimeout"`
	MaxConnLifetime time.Duration `yaml:"maxConnLifetime"`
	ConnTimeOut     time.Duration `yaml:"connTimeOut"`
	ReadTimeOut     time.Duration `yaml:"readTimeOut"`
	WriteTimeOut    time.Duration `yaml:"writeTimeOut"`
}

func (conf *RedisConf) checkConf() {
	env.CommonSecretChange(prefix, *conf, conf)

	if conf.MaxIdle == 0 {
		conf.MaxIdle = 50
	}
	if conf.MaxActive == 0 {
		conf.MaxActive = 100
	}
	if conf.IdleTimeout == 0 {
		conf.IdleTimeout = 3 * time.Minute
	}
	if conf.MaxConnLifetime == 0 {
		conf.MaxConnLifetime = 10 * time.Minute
	}
	if conf.ConnTimeOut == 0 {
		conf.ConnTimeOut = 1200 * time.Millisecond
	}
	if conf.ReadTimeOut == 0 {
		conf.ReadTimeOut = 1200 * time.Millisecond
	}
	if conf.WriteTimeOut == 0 {
		conf.WriteTimeOut = 1200 * time.Millisecond
	}
}

// 日志打印Do args部分支持的最大长度
type Redis struct {
	// 这个service下对应的每个host对应的连接池
	pool    map[string]*redigo.Pool
	Service string
	conf    RedisConf
}

func InitRedisClient(conf RedisConf) (*Redis, error) {
	conf.checkConf()

	c := &Redis{
		Service: conf.Service,
		pool:    nil,
		conf:    conf,
	}
	return c, nil
}

func (r *Redis) lazyInit(addr string) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:         r.conf.MaxIdle,
		MaxActive:       r.conf.MaxActive,
		IdleTimeout:     r.conf.IdleTimeout,
		MaxConnLifetime: r.conf.MaxConnLifetime,
		Wait:            true,
		Dial: func() (conn redigo.Conn, e error) {
			con, err := redigo.Dial(
				"tcp",
				addr,
				redigo.DialPassword(r.conf.Password),
				redigo.DialConnectTimeout(r.conf.ConnTimeOut),
				redigo.DialReadTimeout(r.conf.ReadTimeOut),
				redigo.DialWriteTimeout(r.conf.WriteTimeOut),
			)
			if err != nil {
				return nil, err
			}
			return con, nil
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (r *Redis) choosePool(ctx *gin.Context) (selectedAddr string, conn redigo.Conn, err error) {
	selectedAddr = r.conf.Addr
	if selectedAddr == "" {
		return selectedAddr, nil, errors.New("no valid address found")
	}

	if p, ok := r.pool[selectedAddr]; ok {
		// 如果已为该host创建过了连接池，直接从连接池里选一个连接
		return selectedAddr, p.Get(), nil
	}
	// 否则为该host初始化连接池
	p := r.lazyInit(selectedAddr)
	if r.pool == nil {
		r.pool = make(map[string]*redigo.Pool)
	}
	r.pool[selectedAddr] = p
	return selectedAddr, p.Get(), nil
}

func (r *Redis) Do(ctx *gin.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	start := time.Now()

	// 根据service随机选一个host对应的连接池中的连接
	addr, conn, err := r.choosePool(ctx)
	if err != nil {
		klog.WarnLogger(ctx, "get empty pool ", klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
		return reply, err
	}

	if conn == nil {
		klog.WarnLogger(ctx, "get invalid connection: ", klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
		return reply, errors.New("get nil connection")
	}

	if err := conn.Err(); err != nil {
		klog.ErrorLogger(ctx, "get connection error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
		return reply, err
	}

	reply, err = conn.Do(commandName, args...)
	if e := conn.Close(); e != nil {
		klog.WarnLogger(ctx, "connection close error: "+e.Error(), klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
	}

	end := time.Now()

	// 执行时间 单位:毫秒
	ralCode := 0
	msg := "redis do success"
	if err != nil {
		ralCode = -1
		msg = fmt.Sprintf("redis do error: %s", err.Error())
		klog.ErrorLogger(ctx, msg, klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
	}

	fields := []klog.Field{
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("prot", "redis"),
		klog.String("service", r.Service),
		klog.String("remoteAddr", addr),
		klog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		klog.Float64("cost", utils.GetRequestCost(start, end)),
		klog.String("command", commandName),
		klog.String("commandVal", utils.JoinArgs(logForRedisValue, args)),
		klog.Int("ralCode", ralCode),
	}

	klog.InfoLogger(ctx, msg, fields...)
	return reply, err
}

func (r *Redis) Close() error {
	for _, p := range r.pool {
		_ = p.Close()
	}
	return nil
}

// todo 这个方法之前只针对一个host返回对应的连接池
func (r *Redis) Stats() (inUseCount, idleCount, activeCount int) {
	var pool *redigo.Pool
	if r.conf.Addr != "" {
		pool = r.pool[r.conf.Addr]
	} else {
		for _, p := range r.pool {
			if p != nil {
				pool = p
				break
			}
		}
	}

	if pool == nil {
		klog.WarnLogger(nil, "[Stats] error, not found pool", klog.String(klog.TopicType, klog.LogNameModule), klog.String("prot", "redis"))
		return inUseCount, idleCount, activeCount
	}
	stats := pool.Stats()
	idleCount = stats.IdleCount
	activeCount = stats.ActiveCount
	inUseCount = activeCount - idleCount
	return inUseCount, idleCount, activeCount
}
