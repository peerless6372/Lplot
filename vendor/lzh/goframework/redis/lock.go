package redis

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

const (
	EXSECONDS       = "EX"
	PXMILLISSECONDS = "PX"
	NOTEXISTS       = "NX"
)

//设置过期时间为秒级的redis分布式锁
func (r *Redis) SetNxByEX(ctx *gin.Context, key string, value interface{}, expire uint64) (bool, error) {
	return r.tryLock(ctx, key, value, expire, EXSECONDS)
}

//设置过期时间为毫秒的redis分布式锁
func (r *Redis) SetNxByPX(ctx *gin.Context, key string, value interface{}, expire uint64) (bool, error) {
	return r.tryLock(ctx, key, value, expire, PXMILLISSECONDS)
}

func (r *Redis) tryLock(ctx *gin.Context, key string, value interface{}, expire uint64, exType string) (bool, error) {
	str := parseToString(value)
	if str == "" {
		return false, errors.New("value is empty")
	}

	_, err := redis.String(r.Do(ctx, "SET", key, str, exType, expire, NOTEXISTS))

	if err == redis.ErrNil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
