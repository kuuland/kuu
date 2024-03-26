package kuu

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"regexp"
	"strings"
	"sync"
)

func ParseJSONConfig(data []byte) *JSONConfig {
	return &JSONConfig{
		parsedPath: new(sync.Map),
		data:       data,
	}
}

type JSONConfig struct {
	parsedPath *sync.Map
	data       []byte
}

func (js *JSONConfig) pathToKeys(path string) []string {
	if v, has := js.parsedPath.Load(path); has {
		return v.([]string)
	}

	var reg *regexp.Regexp

	reg = regexp.MustCompile(`(\[\d+\])`)
	path = reg.ReplaceAllString(path, ".$1.")

	reg = regexp.MustCompile(`\.+`)
	path = reg.ReplaceAllString(path, ".")

	path = strings.Trim(path, ".")

	keys := strings.Split(path, ".")

	js.parsedPath.Store(path, keys)
	return keys
}

func (js *JSONConfig) RawGet(path string) (val []byte, exists bool) {
	if v, _, _, err := jsonparser.Get(js.data, js.pathToKeys(path)...); err == nil {
		exists = true
		val = v
	}
	return
}

func (js *JSONConfig) RawGetInt(path string) (int64, error) {
	return jsonparser.GetInt(js.data, js.pathToKeys(path)...)
}

func (js *JSONConfig) RawGetFloat(path string) (float64, error) {
	return jsonparser.GetFloat(js.data, js.pathToKeys(path)...)
}

func (js *JSONConfig) RawSet(path string, value []byte) error {
	newData, err := jsonparser.Set(js.data, value, js.pathToKeys(path)...)
	if err != nil {
		return err
	}
	js.data = newData
	return nil
}

func (js *JSONConfig) Has(path string) bool {
	_, _, _, err := jsonparser.Get(js.data, js.pathToKeys(path)...)
	return err == nil
}

func (js *JSONConfig) Get(path string, dst any) {
	value, _, _, err := jsonparser.Get(js.data, js.pathToKeys(path)...)
	if err == nil {
		_ = json.Unmarshal(value, dst)
	}
}

func (js *JSONConfig) GetString(path string, defaultValue ...string) (s string) {
	v, err := jsonparser.GetString(js.data, js.pathToKeys(path)...)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return v
}

func (js *JSONConfig) GetInt(path string, defaultValue ...int) int {
	v, err := js.RawGetInt(path)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return int(v)
}

func (js *JSONConfig) GetInt32(path string, defaultValue ...int32) int32 {
	v, err := js.RawGetInt(path)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return int32(v)
}

func (js *JSONConfig) GetInt64(path string, defaultValue ...int64) int64 {
	v, err := js.RawGetInt(path)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return v
}
func (js *JSONConfig) GetFloat32(path string, defaultValue ...float32) float32 {
	v, err := js.RawGetFloat(path)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return float32(v)
}

func (js *JSONConfig) GetFloat64(path string, defaultValue ...float64) (f64 float64) {
	v, err := js.RawGetFloat(path)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return v
}

func (js *JSONConfig) GetBool(path string, defaultValue ...bool) bool {
	v, err := jsonparser.GetBoolean(js.data, js.pathToKeys(path)...)
	if len(defaultValue) > 0 && err != nil {
		return defaultValue[0]
	}
	return v
}

func (js *JSONConfig) SetString(path string, value string) error {
	return js.RawSet(path, []byte(fmt.Sprintf("\"%s\"", value)))
}

func (js *JSONConfig) SetInt(path string, value int) error {
	return js.RawSet(path, []byte(fmt.Sprintf("%d", value)))
}

func (js *JSONConfig) SetInt32(path string, value int32) error {
	return js.SetInt(path, int(value))
}

func (js *JSONConfig) SetInt64(path string, value int64) error {
	return js.SetInt(path, int(value))
}

func (js *JSONConfig) SetFloat32(path string, value float32) error {
	return js.SetFloat64(path, float64(value))
}

func (js *JSONConfig) SetFloat64(path string, value float64) error {
	return js.RawSet(path, []byte(fmt.Sprintf("%f", value)))
}

func (js *JSONConfig) SetBool(path string, value bool) error {
	return js.RawSet(path, []byte(fmt.Sprintf("%t", value)))
}

func (js *JSONConfig) ToData() []byte {
	return js.data
}

func (js *JSONConfig) ToAny(dst any) error {
	return json.Unmarshal(js.data, dst)
}

func (js *JSONConfig) ToString() string {
	return string(js.data)
}
