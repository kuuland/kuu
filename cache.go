package kuu

import (
	"github.com/go-redis/redis"
)

var (
	defaultCachePrefix = "kuu"
	redisClient        *redis.Client
)

func initRedis() {
	rawConfig, exists := C().Get("redis")
	if !exists {
		PANIC(`redis config is required:
{
  "redis": {
	"addr": "localhost:6379",
	"pass": "",
	"db": 0
  }
}
NOTE: "pass" and "db" are optional.
`)
	}
	var (
		addr string
		pass string
		db   int
	)
	if m, ok := rawConfig.(map[string]interface{}); ok {
		if v, ok := m["addr"].(string); ok {
			addr = v
		}
		if v, ok := m["pass"].(string); ok {
			pass = v
		}
		if v, ok := m["db"].(int); ok {
			db = v
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	if addr != "" {
		if _, err := client.Ping().Result(); err == nil {
			connectedPrint("Redis", addr)
			redisClient = client
		} else {
			ERROR(err)
		}
	}
}
