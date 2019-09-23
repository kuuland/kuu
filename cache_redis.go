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
	clusterMode   bool
	client        *redis.Client
	clusterClient *redis.ClusterClient
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
	c := &CacheRedis{}
	if C().GetBool("redis:cluster") {
		var opts redis.ClusterOptions
		C().GetInterface("redis", &opts)
		client := redis.NewClusterClient(&opts)
		if _, err := client.Ping().Result(); err != nil {
			PANIC(err)
		}
		c.clusterMode = true
		c.clusterClient = client
		connectedPrint(Capitalize("redis"), strings.Join(opts.Addrs, ","))
	} else {
		var opts redis.Options
		C().GetInterface("redis", &opts)
		client := redis.NewClient(&opts)
		if _, err := client.Ping().Result(); err != nil {
			PANIC(err)
		}
		c.client = client
		connectedPrint(Capitalize("redis"), opts.Addr)
	}
	return c
}

// SetString
func (c *CacheRedis) SetString(rawKey, val string, expiration ...time.Duration) {
	var (
		key, exp = c.buildKeyAndExp(rawKey, expiration)
		status   *redis.StatusCmd
	)
	if c.clusterMode {
		status = c.clusterClient.Set(key, val, exp)
	} else {
		status = c.client.Set(key, val, exp)
	}
	if err := status.Err(); err != nil {
		ERROR(err)
	}
}

// GetString
func (c *CacheRedis) GetString(rawKey string) (val string) {
	var (
		key = c.buildKey(rawKey)
		cmd *redis.StringCmd
	)
	if c.clusterMode {
		cmd = c.clusterClient.Get(key)
	} else {
		cmd = c.client.Get(key)
	}
	if err := cmd.Err(); err != nil {
		ERROR(err)
	} else {
		val = cmd.Val()
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
		cmd *redis.IntCmd
	)
	if c.clusterMode {
		cmd = c.clusterClient.Incr(key)
	} else {
		cmd = c.client.Incr(key)
	}
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
	var cmd *redis.IntCmd
	if c.clusterMode {
		cmd = c.clusterClient.Del(keys...)
	} else {
		cmd = c.client.Del(keys...)
	}
	if err := cmd.Err(); err != nil {
		ERROR(err)
	}
	return
}

// Close
func (c *CacheRedis) Close() {
	if c.clusterMode {
		if c.clusterClient != nil {
			ERROR(c.clusterClient.Close())
		}
	} else {
		if c.client != nil {
			ERROR(c.client.Close())
		}
	}
}
