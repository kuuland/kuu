package kuu

import (
	"fmt"
	"time"
)

// NewDateSn 创建以日期递增的序列化
func NewDateSn(keyPrefix string, valPrefix string, minLen ...int) string {
	min := 3
	if len(minLen) > 0 && minLen[0] != 0 {
		min = minLen[0]
	}
	date := time.Now().Format("20060102")
	key := fmt.Sprintf("%s_%s", keyPrefix, date)
	val := IncrCache(key)
	return fmt.Sprintf("%s%s%0*d", valPrefix, date, min, val)
}
