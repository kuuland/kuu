package kuu

import (
	"github.com/buger/jsonparser"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	configData     []byte
	configInst     *Config
	parsePathCache sync.Map
	//pairs = make(map[string]interface{})
	//inst  *Config
)

func parseKuuJSON(filePath ...string) (data []byte) {
	var configFile string
	if len(filePath) > 0 && filePath[0] != "" {
		configFile = filePath[0]
	}

	if configFile == "" {
		if v := os.Getenv("CONFIG_FILE"); v != "" {
			configFile = v
		}
	}

	if configFile == "" {
		configFile = "kuu.json"
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return
	}
	data, _ = ioutil.ReadFile(configFile)
	return
}

type Config struct {
	data []byte
}

func mergeConfig(r, n map[string]interface{}) map[string]interface{} {
	if r == nil {
		r = make(map[string]interface{})
	}
	if n == nil {
		n = make(map[string]interface{})
	}
	for k, v := range n {
		r[k] = v
	}
	return r
}

func C(newConfig ...map[string]interface{}) *Config {
	if configInst == nil {
		configData = parseKuuJSON()
		configInst = &Config{data: configData}
	}
	if len(newConfig) > 0 {
		src := make(map[string]interface{})
		_ = json.Unmarshal(configData, &src)
		src = mergeConfig(src, newConfig[0])
		configData, _ = json.Marshal(src)
		configInst.data = configData
	}
	return configInst
}

func (c *Config) ParseKeys(path string) []string {
	if v, has := parsePathCache.Load(path); has {
		return v.([]string)
	}

	var reg *regexp.Regexp

	reg = regexp.MustCompile(`(\[\d+\])`)
	path = reg.ReplaceAllString(path, ".$1.")

	reg = regexp.MustCompile(`\.+`)
	path = reg.ReplaceAllString(path, ".")

	path = strings.Trim(path, ".")

	keys := strings.Split(path, ".")
	parsePathCache.Store(path, keys)

	return keys
}

func (c *Config) Get(path string) (val []byte, exists bool) {
	keys := c.ParseKeys(path)
	if v, _, _, err := jsonparser.Get(c.data, keys...); err == nil {
		exists = true
		val = v
	}
	return
}

func (c *Config) Has(path string) bool {
	keys := c.ParseKeys(path)
	_, _, _, err := jsonparser.Get(c.data, keys...)
	return err == nil
}

// GetInterface returns the value associated with the key.
func (c *Config) GetInterface(path string, out interface{}) {
	keys := c.ParseKeys(path)
	value, _, _, err := jsonparser.Get(c.data, keys...)
	if err == nil {
		_ = json.Unmarshal(value, out)
	}
}

// GetString returns the value associated with the key as a string.
func (c *Config) GetString(path string) (s string) {
	keys := c.ParseKeys(path)
	s, _ = jsonparser.GetString(c.data, keys...)
	return
}

// DefaultGetString returns the value associated with the key as a string.
func (c *Config) DefaultGetString(path string, defaultValue string) string {
	keys := c.ParseKeys(path)
	if v, err := jsonparser.GetString(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt64(path string) int64 {
	keys := c.ParseKeys(path)
	value, _ := jsonparser.GetInt(c.data, keys...)
	return value
}

// GetInt returns the value associated with the key as an integer.
func (c *Config) GetInt(path string) (i int) {
	v := c.GetInt64(path)
	return int(v)
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt32(path string) int32 {
	v := c.GetInt64(path)
	return int32(v)
}

// DefaultGetInt returns the value associated with the key as a integer.
func (c *Config) DefaultGetInt(path string, defaultValue int) int {
	keys := c.ParseKeys(path)
	if v, err := jsonparser.GetInt(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return int(v)
	}
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Config) GetFloat64(path string) (f64 float64) {
	keys := c.ParseKeys(path)
	f64, _ = jsonparser.GetFloat(c.data, keys...)
	return
}

// GetFloat64 returns the value associated with the key as a float32.
func (c *Config) GetFloat32(path string) float32 {
	v := c.GetFloat64(path)
	return float32(v)
}

// DefaultGetInt returns the value associated with the key as a float64.
func (c *Config) DefaultGetFloat64(path string, defaultValue float64) float64 {
	keys := c.ParseKeys(path)
	if v, err := jsonparser.GetFloat(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}

// GetBool returns the value associated with the key as a boolean.
func (c *Config) GetBool(path string) (b bool) {
	keys := c.ParseKeys(path)
	b, _ = jsonparser.GetBoolean(c.data, keys...)
	return
}

// DefaultGetBool returns the value associated with the key as a boolean.
func (c *Config) DefaultGetBool(path string, defaultValue bool) bool {
	keys := c.ParseKeys(path)
	if v, err := jsonparser.GetBoolean(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}
