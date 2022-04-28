package kuu

import (
	"github.com/buger/jsonparser"
	"io/ioutil"
	"os"
	"sync"
)

var (
	configData     []byte
	configInst     *Config
	parsePathCache sync.Map
	fromParamKeys  = make(map[string]bool)
)

func parseKuuJSON(filePath ...string) (data []byte) {
	var configFile string
	if len(filePath) > 0 && filePath[0] != "" {
		configFile = filePath[0]
	}

	if configFile == "" {
		if v := os.Getenv("KUU_CONFIG"); v != "" {
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

func (c *Config) Get(path string) (val []byte, exists bool) {
	keys := ParseJSONPath(path)
	if v, _, _, err := jsonparser.Get(c.data, keys...); err == nil {
		exists = true
		val = v
	}
	return
}

func (c *Config) Has(path string) bool {
	keys := ParseJSONPath(path)
	_, _, _, err := jsonparser.Get(c.data, keys...)
	return err == nil
}

// GetInterface returns the value associated with the key.
func (c *Config) GetInterface(path string, out interface{}) {
	keys := ParseJSONPath(path)
	value, _, _, err := jsonparser.Get(c.data, keys...)
	if err == nil {
		_ = json.Unmarshal(value, out)
	}
}

// GetString returns the value associated with the key as a string.
func (c *Config) GetString(path string) (s string) {
	keys := ParseJSONPath(path)
	s, _ = jsonparser.GetString(c.data, keys...)
	return
}

// DefaultGetString returns the value associated with the key as a string.
func (c *Config) DefaultGetString(path string, defaultValue string) string {
	keys := ParseJSONPath(path)
	if v, err := jsonparser.GetString(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt64(path string) int64 {
	keys := ParseJSONPath(path)
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
	keys := ParseJSONPath(path)
	if v, err := jsonparser.GetInt(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return int(v)
	}
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Config) GetFloat64(path string) (f64 float64) {
	keys := ParseJSONPath(path)
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
	keys := ParseJSONPath(path)
	if v, err := jsonparser.GetFloat(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}

// GetBool returns the value associated with the key as a boolean.
func (c *Config) GetBool(path string) (b bool) {
	keys := ParseJSONPath(path)
	b, _ = jsonparser.GetBoolean(c.data, keys...)
	return
}

// DefaultGetBool returns the value associated with the key as a boolean.
func (c *Config) DefaultGetBool(path string, defaultValue bool) bool {
	keys := ParseJSONPath(path)
	if v, err := jsonparser.GetBoolean(c.data, keys...); err != nil {
		return defaultValue
	} else {
		return v
	}
}

func (c *Config) LoadFromParams(kyes ...string) {
	var params []Param
	DB().Model(&Param{}).Where("code in (?)", kyes).Find(&params)
	for _, param := range params {
		fromParamKeys[param.Code] = true
		INFO("Load Config From param: %s(%s)", param.Code, param.Name)
		if param.Type == "json" {
			var err error
			c.data, err = jsonparser.Set(c.data, []byte(param.Value), "params", param.Code)
			if err != nil {
				ERROR("%s: 加载失败: %w", param.Code, err.Error())
				continue
			}
		}
	}
}
