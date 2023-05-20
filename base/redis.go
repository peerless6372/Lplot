package base

import (
	"github.com/peerless6372/Lplot/redis"
	"github.com/peerless6372/Lplot/utils"
	"time"

	"github.com/peerless6372/gin"
)

// deprecated
// use golib/redis.Conf instead
type RedisConf struct {
	Service      string        `yaml:"service"`
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	MaxIdle      int           `yaml:"maxIdle"`
	MaxActive    int           `yaml:"maxActive"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
	ConnTimeOut  time.Duration `yaml:"connTimeOut"`
	ReadTimeOut  time.Duration `yaml:"readTimeOut"`
	WriteTimeOut time.Duration `yaml:"writeTimeOut"`
}

type RedisClient struct {
	*redis.Redis
}

// deprecated
// use golib/redis.InitRedisClient() instead
func InitRedisClient(conf RedisConf) *RedisClient {
	var newConf redis.RedisConf
	if err := utils.Copy(&newConf, conf); err != nil {
		panic("conf error: " + err.Error())
	}

	r, _ := redis.InitRedisClient(newConf)
	c := &RedisClient{
		Redis: r,
	}
	return c
}

func (r *RedisClient) Do(ctx *gin.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	return r.Redis.Do(ctx, commandName, args...)
}
