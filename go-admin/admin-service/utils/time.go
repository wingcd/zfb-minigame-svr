package utils

import (
	"time"
)

// UnixMilli 获取毫秒级时间戳，兼容旧版本 Go
func UnixMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// UnixMilliFromTime 从指定时间获取毫秒级时间戳
func UnixMilliFromTime(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
