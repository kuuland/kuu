package kuu

import (
	"bytes"
	"encoding/json"
)

// Stringify 格式化
func Stringify(v interface{}, format bool) (ret string) {
	if b, err := json.Marshal(v); err == nil {
		if format {
			var out bytes.Buffer
			if err = json.Indent(&out, b, "", "  "); err == nil {
				ret = string(out.Bytes())
			}
		} else {
			ret = string(b)
		}
	}
	return
}

// Parse 解析
func Parse(v string, recv interface{}) {
	json.Unmarshal([]byte(v), recv)
}
