package kuu

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"strings"
	"time"
)

// CacheRedis
type CacheRedis struct {
	client redis.UniversalClient
}

func (c *CacheRedis) buildKey(key string) string {
	return fmt.Sprintf("%s_%s", GetAppName(), key)
}

func (c *CacheRedis) buildKeyAndExp(key string, expiration []time.Duration) (string, time.Duration) {
	key = c.buildKey(key)
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return key, exp
}

// NewCacheRedis
func NewCacheRedis() *CacheRedis {
	GetAppName()
	var (
		c = &CacheRedis{}
	)
	// 解析配置
	var opts redis.UniversalOptions
	C().GetInterface("redis", &opts)
	// 初始化客户端
	cmd := redis.NewUniversalClient(&opts)
	if _, err := cmd.Ping().Result(); err != nil {
		PANIC(err)
	}
	c.client = cmd
	connectedPrint(Capitalize("redis"), strings.Join(opts.Addrs, ","))
	return c
}

// SetString
func (c *CacheRedis) SetString(rawKey, val string, expiration ...time.Duration) {
	var (
		key, exp = c.buildKeyAndExp(rawKey, expiration)
		status   = c.client.Set(key, val, exp)
	)
	if err := status.Err(); err != nil {
		ERROR(err)
	}
}

// GetString
func (c *CacheRedis) GetString(rawKey string) (val string) {
	var (
		key = c.buildKey(rawKey)
		cmd = c.client.Get(key)
	)
	if err := cmd.Err(); err != nil {
		ERROR(err)
	} else {
		val = cmd.Val()
	}
	return
}

func (c *CacheRedis) keys(pattern string) (values map[string]string) {
	cmd := c.client.Keys(pattern)
	if err := cmd.Err(); err != nil {
		ERROR(err)
		return
	}
	values = make(map[string]string)
	if keys := cmd.Val(); len(keys) > 0 {
		for _, key := range keys {
			values[key] = c.client.Get(key).Val()
		}
	}
	return
}

// HasPrefix
func (c *CacheRedis) HasPrefix(rawKey string) map[string]string {
	pattern := c.buildKey(rawKey)
	if !strings.HasSuffix(pattern, "*") {
		pattern = fmt.Sprintf("%s*", pattern)
	}
	return c.keys(pattern)
}

// HasSuffix
func (c *CacheRedis) HasSuffix(rawKey string) map[string]string {
	pattern := c.buildKey(rawKey)
	if !strings.HasPrefix(pattern, "*") {
		pattern = fmt.Sprintf("*%s", pattern)
	}
	return c.keys(pattern)
}

// Contains
func (c *CacheRedis) Contains(rawKey string) map[string]string {
	pattern := c.buildKey(rawKey)
	if !strings.HasPrefix(pattern, "*") {
		pattern = fmt.Sprintf("*%s", pattern)
	}
	if !strings.HasSuffix(pattern, "*") {
		pattern = fmt.Sprintf("%s*", pattern)
	}
	return c.keys(pattern)
}

// Search
func (c *CacheRedis) Search(rawMatch string, filter func(string, string) bool) (values map[string]string) {
	if rawMatch == "" {
		rawMatch = "*"
	}
	if !strings.Contains(rawMatch, "*") {
		rawMatch = fmt.Sprintf("*%s*", rawMatch)
	}
	var (
		match = c.buildKey(rawMatch)
		iter  = c.client.Scan(0, match, 0).Iterator()
	)
	if err := iter.Err(); err != nil {
		ERROR(err)
		return
	}
	values = make(map[string]string)
	for iter.Next() {
		key := iter.Val()
		value := c.client.Get(key).Val()
		if filter(key, value) {
			values[key] = value
		}
	}
	return
}

// SetInt
func (c *CacheRedis) SetInt(rawKey string, val int, expiration ...time.Duration) {
	c.SetString(rawKey, strconv.Itoa(val), expiration...)

}

// GetInt
func (c *CacheRedis) GetInt(key string) (val int) {
	if v, err := strconv.Atoi(c.GetString(key)); err == nil {
		val = v
	}
	return
}

// Incr
func (c *CacheRedis) Incr(rawKey string) (val int) {
	var (
		key = c.buildKey(rawKey)
		cmd = c.client.Incr(key)
	)
	if err := cmd.Err(); err != nil {
		ERROR(err)
	} else {
		val = int(cmd.Val())
	}
	return
}

// Del
func (c *CacheRedis) Del(keys ...string) {
	for index, key := range keys {
		keys[index] = c.buildKey(key)
	}
	cmd := c.client.Del(keys...)
	if err := cmd.Err(); err != nil {
		ERROR(err)
	}
}

// DelLike
func (c *CacheRedis) DelLike(keys ...string) {
	if len(keys) == 0 {
		return
	}
	var delKeys []string
	_ = c.Search("", func(key string, value string) bool {
		for _, k := range keys {
			if strings.Contains(key, k) {
				delKeys = append(delKeys, key)
				return true
			}
		}
		return false
	})
	c.Del(delKeys...)
}

// Close
func (c *CacheRedis) Close() {
	if c.client != nil {
		ERROR(c.client.Close())
	}
}
