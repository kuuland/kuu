package kuu

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

var (
	pairs map[string]interface{}
	inst  *Config
)

type Config struct {
	Keys map[string]interface{}
}

func C() *Config {
	if inst == nil {
		inst = &Config{Keys: pairs}
	}
	return inst
}

func (c *Config) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
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
func (c *Config) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
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
