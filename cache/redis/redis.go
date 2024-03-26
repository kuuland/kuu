package redis

import (
	"context"
	"encoding/json"
	"github.com/kuuland/kuu/v3"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
	"time"
)

var (
	globalKeyPrefix string
	tokenKeyPrefix  string
)

func init() {
	globalKeyPrefix = os.Getenv(kuu.ConfigCacheGlobalKeyPrefix)
	tokenKeyPrefix = os.Getenv(kuu.ConfigCacheTokenKeyPrefix)
}

func New(config string) kuu.Cache {
	var opts redis.UniversalOptions
	if err := json.Unmarshal([]byte(config), &opts); err != nil {
		kuu.Panicln(err)
	}
	client := redis.NewUniversalClient(&opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		kuu.Panicln(err)
	}
	kuu.Infoln("Redis is connected!")
	return &cache{
		client: client,
	}
}

func (c *cache) Publish(ctx context.Context, channel string, message any) (val int64, err error) {
	return c.client.Publish(ctx, channel, message).Result()
}

func (c *cache) Subscribe(ctx context.Context, channels ...string) <-chan *kuu.ChannelMessage {
	msg := make(chan *kuu.ChannelMessage)
	pubsub := c.client.Subscribe(ctx, channels...)
	ch := pubsub.Channel()

	go func() {
		defer func() {
			close(msg) // Close the channel when the goroutine exits
			_ = pubsub.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return // Exit the goroutine if the context is canceled
			case raw, ok := <-ch:
				if !ok {
					return // Exit if the channel is closed
				}

				val := &kuu.ChannelMessage{
					Channel: raw.Channel,
					Pattern: raw.Pattern,
					Payload: raw.Payload,
					// Copy the PayloadSlice to avoid direct reference
					PayloadSlice: append([]string{}, raw.PayloadSlice...),
				}
				msg <- val
			}
		}
	}()

	return msg
}

func (c *cache) BuildKey(key string) string {
	return c.buildKey(key, globalKeyPrefix)
}

func (c *cache) BuildTokenKey(key string) string {
	return c.buildKey(key, tokenKeyPrefix)
}

type cache struct {
	client redis.UniversalClient
}

func (c *cache) Incr(rawKey string, d time.Duration, ctx ...context.Context) (int64, error) {
	key := c.buildKey(rawKey, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	v, err := c.client.Incr(cc, key).Result()
	if err == nil || d > 0 {
		_, err = c.Expire(rawKey, d, cc)
	}
	return v, err
}

func (c *cache) IncrBy(rawKey string, step int64, d time.Duration, ctx ...context.Context) (int64, error) {
	key := c.buildKey(rawKey, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	v, err := c.client.IncrBy(cc, key, step).Result()

	if err == nil || d > 0 {
		_, err = c.Expire(rawKey, d, cc)
	}
	return v, err
}

func (c *cache) IncrByFloat(rawKey string, step float64, d time.Duration, ctx ...context.Context) (float64, error) {
	key := c.buildKey(rawKey, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	v, err := c.client.IncrByFloat(cc, key, step).Result()

	if err == nil || d > 0 {
		_, err = c.Expire(rawKey, d, cc)
	}
	return v, err
}

func (c *cache) Decr(rawKey string, d time.Duration, ctx ...context.Context) (int64, error) {
	key := c.buildKey(rawKey, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	v, err := c.client.Decr(cc, key).Result()

	if err == nil || d > 0 {
		_, err = c.Expire(rawKey, d, cc)
	}
	return v, err
}

func (c *cache) DecrBy(rawKey string, step int64, d time.Duration, ctx ...context.Context) (int64, error) {
	key := c.buildKey(rawKey, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	v, err := c.client.DecrBy(cc, key, step).Result()

	if err == nil || d > 0 {
		_, err = c.Expire(rawKey, d, cc)
	}
	return v, err
}

func (c *cache) Expire(key string, duration time.Duration, ctx ...context.Context) (bool, error) {
	key = c.buildKey(key, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	return c.client.Expire(cc, key, duration).Result()
}

func (c *cache) RawClient() any {
	return c.client
}

func (c *cache) RawSet(key, value string, duration time.Duration, ctx ...context.Context) error {
	return c.set(false, key, value, duration, ctx...)
}

func (c *cache) RawGet(key string, ctx ...context.Context) (string, error) {
	return c.get(false, key, ctx...)
}

func (c *cache) Del(keys []string, ctx ...context.Context) (int64, error) {
	return c.del(true, keys, ctx...)
}

func (c *cache) RawDel(keys []string, ctx ...context.Context) (int64, error) {
	return c.del(false, keys, ctx...)
}

func (c *cache) Keys(pattern string, ctx ...context.Context) ([]string, error) {
	pattern = c.buildKey(pattern, globalKeyPrefix)
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	ss, err := c.client.Keys(cc, pattern).Result()
	if err == redis.Nil {
		return ss, nil
	}
	return ss, err
}

func (c *cache) SetToken(key string, value string, d time.Duration, ctx ...context.Context) error {
	key = c.buildKey(key, tokenKeyPrefix)
	return c.Set(key, value, d, ctx...)
}

func (c *cache) GetToken(key string, ctx ...context.Context) (string, error) {
	key = c.buildKey(key, tokenKeyPrefix)
	return c.Get(key, ctx...)
}

func (c *cache) Set(key, value string, duration time.Duration, ctx ...context.Context) error {
	return c.set(true, key, value, duration, ctx...)
}

func (c *cache) Get(key string, ctx ...context.Context) (string, error) {
	return c.get(true, key, ctx...)
}

func (c *cache) buildKey(rawKey string, prefix ...string) string {
	prefix = append(prefix, rawKey)
	return strings.Join(prefix, "")
}

func (c *cache) set(autoSetGlobalKeyPrefix bool, key, value string, duration time.Duration, ctx ...context.Context) error {
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	if autoSetGlobalKeyPrefix {
		key = c.buildKey(key, globalKeyPrefix)
	}
	return c.client.Set(cc, key, value, duration).Err()
}

func (c *cache) del(autoSetGlobalKeyPrefix bool, keys []string, ctx ...context.Context) (int64, error) {
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	if autoSetGlobalKeyPrefix {
		for i, key := range keys {
			keys[i] = c.buildKey(key, globalKeyPrefix)
		}
	}
	return c.client.Del(cc, keys...).Result()
}

func (c *cache) get(autoSetGlobalKeyPrefix bool, key string, ctx ...context.Context) (string, error) {
	var cc context.Context
	if len(ctx) > 0 {
		cc = ctx[0]
	} else {
		cc = context.Background()
	}
	if autoSetGlobalKeyPrefix {
		key = c.buildKey(key, globalKeyPrefix)
	}
	s, err := c.client.Get(cc, key).Result()
	if err == redis.Nil {
		return s, nil
	}
	return s, err
}
