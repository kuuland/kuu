package kuu

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var (
	pairs map[string]interface{}
	inst  *Config
)

// Parse config file when init
func init() {
	filePath := os.Getenv("CONFIG_FILE")
	if IsBlank(filePath) {
		filePath = "kuu.json"
	}
	pairs = make(map[string]interface{})
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		ERROR(err)
	} else {
		err := json.Unmarshal(data, &pairs)
		if err != nil {
			ERROR(err)
		}
	}
}

type Config struct {
	Keys map[string]interface{}
}

func C(newConfig ...map[string]interface{}) *Config {
	if len(newConfig) > 0 {
		for k, v := range newConfig[0] {
			pairs[k] = v
		}
		inst = nil
	}
	if inst == nil {
		inst = &Config{Keys: pairs}
	}
	return inst
}

func (c *Config) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

// GetInterface returns the value associated with the key.
func (c *Config) GetInterface(key string) interface{} {
	if val, ok := c.Get(key); ok && val != nil {
		return val
	}
	return nil
}

// GetString returns the value associated with the key as a string.
func (c *Config) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Config) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Config) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt32(key string) (i32 int32) {
	if val, ok := c.Get(key); ok && val != nil {
		i32, _ = val.(int32)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float32.
func (c *Config) GetFloat32(key string) (f32 float32) {
	if val, ok := c.Get(key); ok && val != nil {
		f32, _ = val.(float32)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Config) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}
