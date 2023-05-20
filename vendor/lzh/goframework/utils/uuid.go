package utils

import (
	"strconv"

	"github.com/sony/sonyflake"
)

/*
sonyflake 实现方式：
0 - 0000000  00000000  00000000  00000000  00000000 - 00000000 - 00000000 00000000
    |              时间戳,精确到10毫秒                | 序列号    | 16位机器号,默认用私有ip后两位|
1台机器1秒最多生成25600个uuid
不同网段的机器(私有ip后两位相同)可能会生成相同的uuid
遇到时间回拨会等待

*/
var flake = sonyflake.NewSonyflake(sonyflake.Settings{})

func GetUuidString() string {
	return strconv.FormatUint(GetUuidUInt64(), 10)
}

func GetUuidUInt64() uint64 {
	uid, _ := flake.NextID()
	return uid
}
