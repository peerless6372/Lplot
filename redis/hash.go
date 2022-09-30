package redis

import (
	"Lplot/klog"
	"errors"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (r *Redis) HSet(ctx *gin.Context, key, field string, val interface{}) (int, error) {
	valStr := parseToString(val)
	return redis.Int(r.Do(ctx, "HSET", key, field, valStr))
}

func (r *Redis) HGet(ctx *gin.Context, key, field string) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "HGET", key, field)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) HMGet(ctx *gin.Context, key string, fields ...string) ([][]byte, error) {
	//1.初始化返回结果
	res := make([][]byte, 0, len(fields))
	var resErr error
	//2.将多个key分批获取（每次32个）
	pageNum := int(math.Ceil(float64(len(fields)) / float64(_CHUNK_SIZE)))
	for i := 0; i < pageNum; i++ {
		//2.1创建分批切片 []string
		var end int
		if i == (pageNum - 1) {
			end = len(fields)
		} else {
			end = (i + 1) * _CHUNK_SIZE
		}
		chunk := fields[i*_CHUNK_SIZE : end]
		//2.2分批切片的类型转换 => [][]byte
		chunkLength := len(chunk)
		fieldList := make([]interface{}, 0, chunkLength)
		for _, v := range chunk {
			fieldList = append(fieldList, v)
		}
		cacheRes, err := redis.ByteSlices(r.Do(ctx, "HMGET", redis.Args{}.Add(key).AddFlat(fieldList)...))
		if err != nil {
			for i := 0; i < chunkLength; i++ {
				res = append(res, nil)
			}
			klog.Warn(nil, "cache_mget_error: ", err)
			continue
		} else {
			res = append(res, cacheRes...)
		}
	}
	return res, resErr
}

// HMSet 将一个map存到Redis hash
func (r *Redis) HMSet(ctx *gin.Context, key string, fvmap map[string]interface{}) error {
	_, err := r.Do(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(fvmap)...)
	return err
}

func (r *Redis) HKeys(ctx *gin.Context, key string) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "HKEYS", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) HGetAll(ctx *gin.Context, key string) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "HGETALL", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) HLen(ctx *gin.Context, key string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "HLEN", key)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

func (r *Redis) HVals(ctx *gin.Context, key string) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "HVALS", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) HIncrBy(ctx *gin.Context, key, field string, value int64) (int64, error) {
	return redis.Int64(r.Do(ctx, "HINCRBY", key, field, value))
}

func (r *Redis) HExists(ctx *gin.Context, key string, field string) (bool, error) {
	if res, err := redis.Bool(r.Do(ctx, "HEXISTS", key, field)); err == redis.ErrNil {
		return false, nil
	} else {
		return res, err
	}
}

func (r *Redis) HDel(ctx *gin.Context, key string, fields ...string) (int64, error) {
	args := packArgs(key, fields)
	if res, err := redis.Int64(r.Do(ctx, "HDEL", args...)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 基于游标的迭代器，每次被调用会返回新的游标，在下次迭代时，需要使用这个新游标作为游标参数，以此来延续之前的迭代过程
// param: key
// param: cursor 游标 传""表示开始新迭代
// param: count 每次迭代返回元素的最大值，limit hint，实际数量并不准确=count
// param: pattern 模式参数，符合glob风格  ? (一个字符) * （任意个字符） [] (匹配其中的任意一个字符)  \x (转义字符)
// return: 新的cursor，filed-value map  当返回""，空map时，表示迭代已结束
func (r *Redis) HScan(ctx *gin.Context, key string, cursor uint64, pattern string, count int) (uint64, map[string][]byte, error) {
	args := packArgs(key, cursor)
	if pattern != "" {
		args = append(args, "MATCH", pattern)
	}
	if count > 0 {
		args = append(args, "COUNT", count)
	}
	values, err := redis.Values(r.Do(ctx, "HSCAN", args...))
	if err == redis.ErrNil {
		return 0, nil, nil
	} else if err != nil {
		return 0, nil, err
	}
	return parseScanResults(values)
}

func parseScanResults(results []interface{}) (uint64, map[string][]byte, error) {
	if len(results) != 2 {
		return 0, nil, errors.New("hscan err length")
	}

	cursorIndex, err := strconv.ParseInt(string(results[0].([]byte)), 10, 64)
	if err != nil {
		return 0, nil, err
	}
	result := make(map[string][]byte)
	scanData := results[1].([]interface{})
	for i := 0; i < len(scanData); i = i + 2 {
		key := string(scanData[i].([]byte))
		result[key] = scanData[i+1].([]byte)
	}
	return uint64(cursorIndex), result, nil
}
