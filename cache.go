package kuu

import (
	"context"
	"time"
)

type Cache interface {
	Set(key, value string, d time.Duration, ctx ...context.Context) error

	BuildKey(key string) string
	BuildTokenKey(key string) string
	Get(key string, ctx ...context.Context) (string, error)
	Expire(key string, d time.Duration, ctx ...context.Context) (bool, error)
	Incr(key string, d time.Duration, ctx ...context.Context) (int64, error)
	IncrBy(key string, step int64, d time.Duration, ctx ...context.Context) (int64, error)
	IncrByFloat(key string, step float64, d time.Duration, ctx ...context.Context) (float64, error)
	Decr(key string, d time.Duration, ctx ...context.Context) (int64, error)
	DecrBy(key string, step int64, d time.Duration, ctx ...context.Context) (int64, error)
	Del(keys []string, ctx ...context.Context) (int64, error)
	Keys(pattern string, ctx ...context.Context) ([]string, error)
	SetToken(key string, value string, d time.Duration, ctx ...context.Context) error
	GetToken(key string, ctx ...context.Context) (string, error)

	RawSet(key, value string, d time.Duration, ctx ...context.Context) error
	RawGet(key string, ctx ...context.Context) (string, error)
	RawDel(keys []string, ctx ...context.Context) (int64, error)
	RawClient() any

	Subscribe(ctx context.Context, channels ...string) <-chan *ChannelMessage
	Publish(ctx context.Context, channel string, message any) (val int64, err error)
}

type ChannelMessage struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}
