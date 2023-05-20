package redis

import (
	"math"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

const (
	_CHUNK_SIZE = 32
)

func (r *Redis) Get(ctx *gin.Context, key string) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "GET", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) MGet(ctx *gin.Context, keys ...string) [][]byte {
	//1.初始化返回结果
	res := make([][]byte, 0, len(keys))

	//2.将多个key分批获取（每次32个）
	pageNum := int(math.Ceil(float64(len(keys)) / float64(_CHUNK_SIZE)))
	for n := 0; n < pageNum; n++ {
		//2.1创建分批切片 []string
		var end int
		if n != (pageNum - 1) {
			end = (n + 1) * _CHUNK_SIZE
		} else {
			end = len(keys)
		}
		chunk := keys[n*_CHUNK_SIZE : end]
		//2.2分批切片的类型转换 => []interface{}
		chunkLength := len(chunk)
		keyList := make([]interface{}, 0, chunkLength)
		for _, v := range chunk {
			keyList = append(keyList, v)
		}
		cacheRes, err := redis.ByteSlices(r.Do(ctx, "MGET", keyList...))
		if err != nil {
			for i := 0; i < len(keyList); i++ {
				res = append(res, nil)
			}
		} else {
			res = append(res, cacheRes...)
		}
	}
	return res
}

func (r *Redis) MSet(ctx *gin.Context, values ...interface{}) error {
	_, err := r.Do(ctx, "MSET", values...)
	return err
}

func (r *Redis) Set(ctx *gin.Context, key string, value interface{}, expire ...int64) error {
	var res string
	var err error
	if expire == nil {
		res, err = redis.String(r.Do(ctx, "SET", key, value))
	} else {
		res, err = redis.String(r.Do(ctx, "SET", key, value, "EX", expire[0]))
	}
	if err != nil {
		return err
	} else if strings.ToLower(res) != "ok" {
		return errors.New("set result not OK")
	}
	return nil
}

func (r *Redis) SetEx(ctx *gin.Context, key string, value interface{}, expire int64) error {
	return r.Set(ctx, key, value, expire)
}

func (r *Redis) Append(ctx *gin.Context, key string, value interface{}) (int, error) {
	return redis.Int(r.Do(ctx, "APPEND", key, value))
}

func (r *Redis) Incr(ctx *gin.Context, key string) (int64, error) {
	return redis.Int64(r.Do(ctx, "INCR", key))
}

func (r *Redis) IncrBy(ctx *gin.Context, key string, value int64) (int64, error) {
	return redis.Int64(r.Do(ctx, "INCRBY", key, value))
}

func (r *Redis) IncrByFloat(ctx *gin.Context, key string, value float64) (float64, error) {
	return redis.Float64(r.Do(ctx, "INCRBYFLOAT", key, value))
}

func (r *Redis) Decr(ctx *gin.Context, key string) (int64, error) {
	return redis.Int64(r.Do(ctx, "DECR", key))
}

func (r *Redis) DecrBy(ctx *gin.Context, key string, value int64) (int64, error) {
	return redis.Int64(r.Do(ctx, "DECRBY", key, value))
}
