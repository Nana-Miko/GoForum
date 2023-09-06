package util

import (
	"fmt"
	"time"
)

// CurrentTimeStampStrMilli 获取当前时间戳字符串(毫秒级)
func CurrentTimeStampStrMilli() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// CurrentTimeStampMilli 获取当前时间戳字符串(毫秒级)
func CurrentTimeStampMilli() int64 {
	return time.Now().UnixMilli()
}
