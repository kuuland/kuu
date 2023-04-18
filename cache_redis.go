package kuu

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
	"time"
)

// CacheRedis
type CacheRedis struct {
	client redis.UniversalClient
}

// GetRedisClient
func GetRedisClient() redis.UniversalClient {
	return DefaultCache.(*CacheRedis).client
}

// BuildKey
func BuildKey(keys ...string) string {
	var (
		rawKey    = strings.Join(keys, "_")
		appPrefix = GetAppName()
	)

	if strings.HasPrefix(rawKey, appPrefix) {
		return rawKey
	}

	return fmt.Sprintf("%s_%s", appPrefix, rawKey)
}

func (c *CacheRedis) buildKeyAndExp(key string, expiration []time.Duration) (string, time.Duration) {
	key = BuildKey(key)
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return key, exp
}

// NewCacheRedis
func NewCacheRedis() *CacheRedis {
	return NewCacheRedisWithName("redis")
}

func NewCacheRedisWithName(name string) *CacheRedis {
	GetAppName()
	var (
		c = &CacheRedis{}
	)
	// 解析配置
	var opts redis.UniversalOptions
	C().GetInterface(name, &opts)
	// 初始化客户端
	cmd := redis.NewUniversalClient(&opts)
	if _, err := cmd.Ping(context.Background()).Result(); err != nil {
		PANIC(err)
	}
	c.client = cmd
	connectedPrint(name, strings.Join(opts.Addrs, ","))
	return c
}

// SetString
func (c *CacheRedis) SetString(rawKey, val string, expiration ...time.Duration) {
	var (
		key, exp = c.buildKeyAndExp(rawKey, expiration)
		status   = c.client.Set(context.Background(), key, val, exp)
	)
	if err := status.Err(); err != nil {
		ERROR(err)
	}
}

// GetString
func (c *CacheRedis) GetString(rawKey string) (val string) {
	var (
		key = BuildKey(rawKey)
		cmd = c.client.Get(context.Background(), key)
	)
	return cmd.Val()
}

func (c *CacheRedis) scan(cursor uint64, pattern string, limit int64) (values map[string]string) {
	values = make(map[string]string)
	for len(values) < int(limit) {
		cmd := c.client.Scan(context.Background(), cursor, pattern, limit)
		if err := cmd.Err(); err != nil {
			ERROR(err)
			return
		}

		if keys, nextCur := cmd.Val(); len(keys) > 0 {
			for _, key := range keys {
				values[key] = c.client.Get(context.Background(), key).Val()
			}
			if (limit != 0 && len(values) >= int(limit)) || nextCur == 0 {
				break
			} else {
				cursor = nextCur
			}
		} else {
			break
		}
	}
	return
}

// HasPrefix
func (c *CacheRedis) HasPrefix(rawKey string, limit int) map[string]string {
	pattern := BuildKey(rawKey)
	if !strings.HasSuffix(pattern, "*") {
		pattern = fmt.Sprintf("%s*", pattern)
	}
	return c.scan(0, pattern, int64(limit))
}

// HasSuffix
func (c *CacheRedis) HasSuffix(rawKey string, limit int) map[string]string {
	pattern := BuildKey(rawKey)
	if !strings.HasPrefix(pattern, "*") {
		pattern = fmt.Sprintf("*%s", pattern)
	}
	return c.scan(0, pattern, int64(limit))
}

// Contains
func (c *CacheRedis) Contains(rawKey string, limit int) map[string]string {
	pattern := BuildKey(rawKey)
	if !strings.HasPrefix(pattern, "*") {
		pattern = fmt.Sprintf("*%s", pattern)
	}
	if !strings.HasSuffix(pattern, "*") {
		pattern = fmt.Sprintf("%s*", pattern)
	}
	return c.scan(0, pattern, int64(limit))
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
		key = BuildKey(rawKey)
		cmd = c.client.Incr(context.Background(), key)
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
		keys[index] = BuildKey(key)
	}
	cmd := c.client.Del(context.Background(), keys...)
	if err := cmd.Err(); err != nil {
		ERROR(err)
	}
}

// Close
func (c *CacheRedis) Close() {
	if c.client != nil {
		ERROR(c.client.Close())
	}
}

func (c *CacheRedis) Publish(channel string, message interface{}) error {
	return c.client.Publish(context.Background(), channel, message).Err()
}

func (c *CacheRedis) Subscribe(channels []string, handler func(string, string)) error {
	ps := c.client.Subscribe(context.Background(), channels...)
	if _, err := ps.Receive(context.Background()); err != nil {
		return err
	}
	ch := ps.Channel()
	go func(ch <-chan *redis.Message) {
		for msg := range ch {
			handler(msg.Channel, msg.Payload)
		}
	}(ch)
	return nil
}

func (c *CacheRedis) PSubscribe(patterns []string, handler func(string, string)) error {
	pubsub := c.client.PSubscribe(context.Background(), patterns...)
	if _, err := pubsub.Receive(context.Background()); err != nil {
		return err
	}
	ch := pubsub.Channel()
	go func(ch <-chan *redis.Message) {
		for msg := range ch {
			handler(msg.Channel, msg.Payload)
		}
	}(ch)
	return nil
}

func (c *CacheRedis) HGetAll(key string) map[string]string {
	return c.client.HGetAll(context.Background(), key).Val()
}

func (c *CacheRedis) HGet(key, field string) string {
	return c.client.HGet(context.Background(), key, field).Val()
}

func (c *CacheRedis) HSet(key string, values ...string) {
	var vs []interface{}
	for _, value := range values {
		vs = append(vs, value)
	}
	c.client.HSet(context.Background(), key, vs...)
}
