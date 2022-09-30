package redis

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

// 将一个或多个 member 元素加入到有序集 key 当中，已经存在于集合的 member 元素将更新该元素的 score 值
// param: maps Member-Score集合
// return: 被添加到集合中的新元素的数量，不包括被更新的、已存在的元素
func (r *Redis) ZAdd(ctx *gin.Context, key string, maps map[string]float64) (int64, error) {
	args := packArgs(key)
	for member, score := range maps {
		args = append(args, score, member)
	}
	return redis.Int64(r.Do(ctx, "ZADD", args...))
}

// 返回有序集 key 中，成员 member 的 score 值
// return: score值，若该成员不存在，返回nil
func (r *Redis) ZScore(ctx *gin.Context, key string, member string) (string, error) {
	if res, err := redis.String(r.Do(ctx, "ZSCORE", key, member)); err == redis.ErrNil {
		return "", nil
	} else {
		return res, err
	}
}

// 为有序集 key 的成员 member 的 score 值加上增量 delta
func (r *Redis) ZIncrBy(ctx *gin.Context, key string, delta float64, member string) (float64, error) {
	return redis.Float64(r.Do(ctx, "ZINCRBY", key, delta, member))
}

// 返回有序集 key 的基数
func (r *Redis) ZCard(ctx *gin.Context, key string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZCARD", key)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中， score 值在 min 和 max 之间(默认包括 score 值等于 min 或 max )的成员的数量
func (r *Redis) ZCount(ctx *gin.Context, key, min, max string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZCOUNT", key, min, max)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 对于一个所有成员的分值都相同的有序集合键 key 来说， 这个命令会返回该集合中， 成员介于 min 和 max 范围内的元素数量。
func (r *Redis) ZLexCount(ctx *gin.Context, key, min, max string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZLEXCOUNT", key, min, max)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中，指定区间内的成员。其中成员的位置按 score 值递增(从小到大)来排序
// withscores指定是否返回得分
// return: Score-Member集合
func (r *Redis) ZRange(ctx *gin.Context, key string, start int, stop int, withscores bool) ([][]byte, error) {
	args := []interface{}{key, start, stop}
	if withscores {
		args = append(args, "WITHSCORES")
	}
	if res, err := redis.ByteSlices(r.Do(ctx, "ZRANGE", args...)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中，指定区间内的成员。其中成员的位置按 score 值递增(从大到小)来排序
// return: Score-Member集合
func (r *Redis) ZRevRange(ctx *gin.Context, key string, start int, stop int, withscores bool) ([][]byte, error) {
	args := []interface{}{key, start, stop}
	if withscores {
		args = append(args, "WITHSCORES")
	}
	if res, err := redis.ByteSlices(r.Do(ctx, "ZREVRANGE", args...)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中，所有 score 值介于 min 和 max 之间(包括等于 min 或 max )的成员。
//有序集成员按 score 值递增(从小到大)次序排列。
// withscores指定是否返回得分
//limit 是否分页方法，false返回所有的数据
func (r *Redis) ZRangeByScore(ctx *gin.Context, key, min, max string, withscores, limit bool, offset int, count int) ([][]byte, error) {
	args := []interface{}{key, min, max}
	if withscores {
		args = append(args, "WITHSCORES")
	}
	if limit {
		args = append(args, "LIMIT", offset, count)
	}
	if res, err := redis.ByteSlices(r.Do(ctx, "ZRANGEBYSCORE", args...)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中，所有 score 值介于 min 和 max 之间(包括等于 min 或 max )的成员。有序集成员按 score 值递增(从大到小)次序排列。
// 如："key", "-inf", "(2"
func (r *Redis) ZRevRangeByScore(ctx *gin.Context, key, min, max string, withscores, limit bool, offset int, count int) ([][]byte, error) {
	args := []interface{}{key, max, min}
	if withscores {
		args = append(args, "WITHSCORES")
	}
	if limit {
		args = append(args, "LIMIT", offset, count)
	}
	if res, err := redis.ByteSlices(r.Do(ctx, "ZREVRANGEBYSCORE", args...)); err == redis.ErrNil {
		return nil, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中成员 member 的排名。其中有序集成员按 score 值递增(从小到大)顺序排列。
// 排名以 0 为底，也就是说， score 值最小的成员排名为 0 。
func (r *Redis) ZRank(ctx *gin.Context, key string, member string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZRANK", key, member)); err == redis.ErrNil {
		return -1, nil
	} else {
		return res, err
	}
}

// 返回有序集 key 中成员 member 的排名。其中有序集成员按 score 值递增(从大到小)顺序排列。
// 排名以 0 为底，也就是说， score 值最小的成员排名为 0 。
func (r *Redis) ZRevRank(ctx *gin.Context, key string, member string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZREVRANK", key, member)); err == redis.ErrNil {
		return -1, nil
	} else {
		return res, err
	}
}

// 移除有序集 key 中的一个或多个成员，不存在的成员将被忽略
// return: 被成功移除的成员的数量，不包括被忽略的成员
func (r *Redis) ZRem(ctx *gin.Context, key string, members ...string) (int64, error) {
	args := packArgs(key, members)
	if res, err := redis.Int64(r.Do(ctx, "ZREM", args...)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 移除有序集 key 中，指定排名(rank)区间内的所有成员。
// 区间分别以下标参数 start 和 stop 指出，包含 start 和 stop 在内。
// return: 被移除成员的数量
func (r *Redis) ZRemRangeByRank(ctx *gin.Context, key string, start int, stop int) (int64, error) {
	args := []interface{}{key, start, stop}
	if res, err := redis.Int64(r.Do(ctx, "ZREMRANGEBYRANK", args...)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 移除有序集 key 中，所有 score 值介于 min 和 max 之间(包括等于 min 或 max )的成员。
// 如："key", "-inf", "(2"
// return: 被移除成员的数量
func (r *Redis) ZRemRangeByScore(ctx *gin.Context, key, min, max string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZREMRANGEBYSCORE", key, min, max)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 对于一个所有成员的分值都相同的有序集合键 key 来说， 这个命令会移除该集合中， 成员介于 min 和 max 范围内的所有元素。
// 如："key", "-inf", "(2"
func (r *Redis) ZRemRangeByLex(ctx *gin.Context, key, min, max string) (int64, error) {
	if res, err := redis.Int64(r.Do(ctx, "ZREMRANGEBYLEX", key, min, max)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// ZUnionStore destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
// 计算给定的一个或多个有序集的并集，其中给定 key 的数量必须以 numkeys 参数指定，并将该并集(结果集)储存到 destination 。
func (r *Redis) ZUnionStore(ctx *gin.Context, destination string, keys []string, weights []int, aggregate string) (int64, error) {
	args := packArgs(destination, len(keys), keys)
	if weights != nil && len(weights) > 0 {
		args = append(args, "WEIGHTS")
		for _, w := range weights {
			args = append(args, w)
		}
	}
	if aggregate != "" {
		args = append(args, "AGGREGATE", aggregate)
	}
	if res, err := redis.Int64(r.Do(ctx, "ZUNIONSTORE", args...)); err == redis.ErrNil {
		return 0, nil
	} else {
		return res, err
	}
}

// 计算给定的一个或多个有序集的交集，其中给定 key 的数量必须以 numkeys 参数指定，并将该交集(结果集)储存到 destination 。
func (r *Redis) ZInterStore(ctx *gin.Context, destination string, keys []string, weights []int, aggregate string) (int64, error) {
	args := packArgs(destination, len(keys), keys)
	if weights != nil && len(weights) > 0 {
		args = append(args, "WEIGHTS")
		for _, w := range weights {
			args = append(args, w)
		}
	}
	if aggregate != "" {
		args = append(args, "AGGREGATE", aggregate)
	}
	if res, err := redis.Int64(r.Do(ctx, "ZINTERSTORE", args...)); err == redis.ErrNil {
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
// return: 新的cursor，score-member pair  当返回""，空map时，表示迭代已结束
func (r *Redis) ZScan(ctx *gin.Context, key string, cursor uint64, pattern string, count int) (uint64, []string, error) {
	args := packArgs(key, cursor)
	if pattern != "" {
		args = append(args, "MATCH", pattern)
	}
	if count > 0 {
		args = append(args, "COUNT", count)
	}
	values, err := redis.Values(r.Do(ctx, "ZSCAN", args...))
	if err != nil {
		return 0, nil, err
	}
	var items []string
	_, err = redis.Scan(values, &cursor, &items)
	if err != nil {
		return 0, nil, err
	}
	return cursor, items, nil
}
