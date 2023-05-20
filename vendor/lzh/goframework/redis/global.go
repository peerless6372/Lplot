package redis

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

func (r *Redis) Expire(ctx *gin.Context, key string, time int64) (bool, error) {
	return redis.Bool(r.Do(ctx, "EXPIRE", key, time))
}

func (r *Redis) Exists(ctx *gin.Context, key string) (bool, error) {
	return redis.Bool(r.Do(ctx, "EXISTS", key))
}

func (r *Redis) Del(ctx *gin.Context, keys ...interface{}) (int64, error) {
	return redis.Int64(r.Do(ctx, "DEL", keys...))
}

func (r *Redis) Ttl(ctx *gin.Context, key string) (int64, error) {
	return redis.Int64(r.Do(ctx, "TTL", key))
}

func (r *Redis) Pttl(ctx *gin.Context, key string) (int64, error) {
	return redis.Int64(r.Do(ctx, "PTTL", key))
}
