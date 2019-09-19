package kuu

import (
	"encoding/binary"
	"github.com/mojocn/base64Captcha"
	"time"
)

// DefaultCacher
var DefaultCacher Cacher

// Cacher
type Cacher interface {
	SetString(string, string, ...time.Duration)
	GetString(string) string
	SetInt(string, int, ...time.Duration)
	GetInt(string) int
	Incr(string) int
	Del(...string)
	Close()
}

func init() {
	if _, exists := C().Get("redis"); exists {
		// 初始化redis
		DefaultCacher = NewCacherRedis()
	} else {
		// 初始化bolt
		DefaultCacher = NewCacherBolt()
	}
	// 初始化验证码存储器
	base64Captcha.SetCustomStore(&captchaStore{})
}

func releaseCacheDB() {
	if DefaultCacher != nil {
		DefaultCacher.Close()
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
	if DefaultCacher != nil {
		DefaultCacher.SetString(key, val, expiration...)
	}
}

// GetCacheString
func GetCacheString(key string) (val string) {
	if DefaultCacher != nil {
		val = DefaultCacher.GetString(key)
	}
	return
}

// SetCacheInt
func SetCacheInt(key string, val int, expiration ...time.Duration) {
	if DefaultCacher != nil {
		DefaultCacher.SetInt(key, val, expiration...)
	}
}

// GetCacheInt
func GetCacheInt(key string) (val int) {
	if DefaultCacher != nil {
		val = DefaultCacher.GetInt(key)
	}
	return
}

// IncrCache
func IncrCache(key string) (val int) {
	if DefaultCacher != nil {
		val = DefaultCacher.Incr(key)
	}
	return
}

// DelCache
func DelCache(keys ...string) {
	if DefaultCacher != nil {
		DefaultCacher.Del(keys...)
	}
	return
}
