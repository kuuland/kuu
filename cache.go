package kuu

import (
	"encoding/binary"
	"time"
)

// DefaultCache
var DefaultCache Cache

// Cache todo
type Cache interface {
	SetString(string, string, ...time.Duration)
	HasPrefix(string, int) map[string]string
	HasSuffix(string, int) map[string]string
	Contains(string, int) map[string]string
	GetString(string) string
	SetInt(string, int, ...time.Duration)
	GetInt(string) int
	Incr(string) int
	Del(...string)
	Close()

	HGetAll(string) map[string]string
	HGet(string, string) string
	HSet(string, ...string)

	Publish(channel string, message interface{}) error
	Subscribe(channels []string, handler func(string, string)) error
	PSubscribe(patterns []string, handler func(string, string)) error
}

func init() {
	if C().Has("redis") {
		// 初始化redis
		DefaultCache = NewCacheRedis()
	} else {
		// 初始化bolt
		DefaultCache = NewCacheBolt()
	}
	_ = DefaultCache.Subscribe([]string{intlMessagesChangedChannel}, func(c string, d string) {
		ReloadIntlMessages()
	})
}

func releaseCacheDB() {
	if DefaultCache != nil {
		DefaultCache.Close()
	}
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) (v int) {
	if b != nil {
		v = int(binary.BigEndian.Uint64(b))
	}
	return
}

// SetCacheString
func SetCacheString(key, val string, expiration ...time.Duration) {
	if DefaultCache != nil {
		DefaultCache.SetString(key, val, expiration...)
	}
}

// GetCacheString
func GetCacheString(key string) (val string) {
	if DefaultCache != nil {
		val = DefaultCache.GetString(key)
	}
	return
}

// SetCacheInt
func SetCacheInt(key string, val int, expiration ...time.Duration) {
	if DefaultCache != nil {
		DefaultCache.SetInt(key, val, expiration...)
	}
}

// GetCacheInt
func GetCacheInt(key string) (val int) {
	if DefaultCache != nil {
		val = DefaultCache.GetInt(key)
	}
	return
}

// IncrCache
func IncrCache(key string) (val int) {
	if DefaultCache != nil {
		val = DefaultCache.Incr(key)
	}
	return
}

// HasPrefixCache
func HasPrefixCache(key string, limit int) (val map[string]string) {
	if DefaultCache != nil {
		val = DefaultCache.HasPrefix(key, limit)
	}
	return
}

// HasSuffixCache
func HasSuffixCache(key string, limit int) (val map[string]string) {
	if DefaultCache != nil {
		val = DefaultCache.HasSuffix(key, limit)
	}
	return
}

// ContainsCache
func ContainsCache(key string, limit int) (val map[string]string) {
	if DefaultCache != nil {
		val = DefaultCache.Contains(key, limit)
	}
	return
}

// DelCache
func DelCache(keys ...string) {
	if DefaultCache != nil {
		DefaultCache.Del(keys...)
	}
	return
}

// PublishCache
func PublishCache(channel string, message interface{}) error {
	if DefaultCache != nil {
		return DefaultCache.Publish(channel, message)
	}
	return nil
}

// SubscribeCache
func SubscribeCache(channels []string, handler func(string, string)) error {
	if DefaultCache != nil {
		return DefaultCache.Subscribe(channels, handler)
	}
	return nil
}

// PSubscribeCache
func PSubscribeCache(patterns []string, handler func(string, string)) error {
	if DefaultCache != nil {
		return DefaultCache.PSubscribe(patterns, handler)
	}
	return nil
}
