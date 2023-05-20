package redis

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gomodule/redigo/redis"
)

// return: 执行命令后，list的长度
func (r *Redis) LPush(ctx *gin.Context, key string, members ...interface{}) (int, error) {
	return redis.Int(r.Do(ctx, "LPUSH", redis.Args{}.Add(key).AddFlat(members)...))
}

// 将值 value 插入到列表 key 的表头，当且仅当 key 存在并且是一个列表。
// return: 命令执行后，list的长度
func (r *Redis) LPushX(ctx *gin.Context, key string, member interface{}) (int, error) {
	return redis.Int(r.Do(ctx, "LPUSHX", key, member))
}

func (r *Redis) RPush(ctx *gin.Context, key string, members ...interface{}) (int, error) {
	return redis.Int(r.Do(ctx, "RPUSH", redis.Args{}.Add(key).AddFlat(members)...))
}

func (r *Redis) RPushX(ctx *gin.Context, key string, member interface{}) (int, error) {
	return redis.Int(r.Do(ctx, "RPUSHX", key, member))
}

// 移除并返回列表 key 的头元素
// return: 列表的头元素，当 key 不存在时，返回 nil,nil
func (r *Redis) LPop(ctx *gin.Context, key string) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "LPOP", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

func (r *Redis) RPop(ctx *gin.Context, key string) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "RPOP", key)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 将列表 source 中的最后一个元素(尾元素)弹出，并返回
// 将 source 弹出的元素插入到列表 destination ，作为 destination 列表的的头元素。
func (r *Redis) RPopLPush(ctx *gin.Context, sourceKey string, destKey string) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "RPOPLPUSH", sourceKey, destKey)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 根据参数 count 的值，移除列表中与参数 value 相等的元素
// count 的值可以是以下几种：
//     count > 0 : 从表头开始向表尾搜索，移除与 value 相等的元素，数量为 count 。
//     count < 0 : 从表尾开始向表头搜索，移除与 value 相等的元素，数量为 count 的绝对值。
//     count = 0 : 移除表中所有与 value 相等的值。
// return: 被移除元素的数量
func (r *Redis) LRem(ctx *gin.Context, key string, count int, value interface{}) (int, error) {
	if res, err := redis.Int(r.Do(ctx, "LREM", key, count, value)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

func (r *Redis) LLen(ctx *gin.Context, key string) (int, error) {
	if res, err := redis.Int(r.Do(ctx, "LLEN", key)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 返回列表 key 中，下标为 index 的元素
func (r *Redis) LIndex(ctx *gin.Context, key string, index int) ([]byte, error) {
	if res, err := redis.Bytes(r.Do(ctx, "LINDEX", key, index)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 将值 value 插入到列表 key 当中，位于值 pivot 之前或之后
func (r *Redis) LInsert(ctx *gin.Context, key string, before bool, pivot interface{}, value interface{}) (int, error) {
	if before {
		if res, err := redis.Int(r.Do(ctx, "LINSERT", key, "BEFORE", pivot, value)); err == redis.ErrNil {
			return 0, nil
		} else {
			return res, err
		}
	} else {
		if res, err := redis.Int(r.Do(ctx, "LINSERT", key, "AFTER", pivot, value)); err == redis.ErrNil {
			return 0, nil
		} else {
			return res, err
		}
	}
}

// 将列表 key 下标为 index 的元素的值设置为 value
func (r *Redis) LSet(ctx *gin.Context, key string, index int, value interface{}) (bool, error) {
	if res, err := redis.String(r.Do(ctx, "LSET", key, index, value)); err == redis.ErrNil {
		return false, nil
	} else if err != nil || strings.ToLower(res) != "ok" {
		return false, err
	} else {
		return true, err
	}
}

// 返回列表 key 中指定区间内的元素，区间以偏移量 start 和 stop 指定。包含 stop 位置的元素
func (r *Redis) LRange(ctx *gin.Context, key string, start int, stop int) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "LRANGE", key, start, stop)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。
func (r *Redis) LTrim(ctx *gin.Context, key string, start int, stop int) (bool, error) {
	if res, err := redis.String(r.Do(ctx, "LTRIM", key, start, stop)); err == redis.ErrNil {
		return false, nil
	} else if err != nil || strings.ToLower(res) != "ok" {
		return false, err
	} else {
		return true, err
	}
}

// 当给定列表内没有任何元素可供弹出的时候，连接将被 BLPOP 命令阻塞，直到等待超时或发现可弹出元素为止。
// timout单位为:秒 设置为0表示阻塞时间无限期延长
// return: 一个含有两个元素的列表，第一个元素是被弹出元素所属的 key ，第二个元素是被弹出元素的值
func (r *Redis) BLPop(ctx *gin.Context, key string, timeout int64) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "BLPOP", key, timeout)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 当给定列表内没有任何元素可供弹出的时候，连接将被 BRPOP 命令阻塞，直到等待超时或发现可弹出元素为止。
// timout单位为:秒 设置为0表示阻塞时间无限期延长
// return: 一个含有两个元素的列表，第一个元素是被弹出元素所属的 key ，第二个元素是被弹出元素的值
func (r *Redis) BRPop(ctx *gin.Context, key string, timeout int64) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "BRPOP", key, timeout)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// timeout单位为:秒 设置为0表示阻塞时间无限期延长
func (r *Redis) BRPopLPush(ctx *gin.Context, sourceKey string, destKey string, timeout int64) ([][]byte, error) {
	if res, err := redis.ByteSlices(r.Do(ctx, "BRPOPLPUSH", sourceKey, destKey, timeout)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}
