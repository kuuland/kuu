package kuu

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

const (
	ASCIISignerAlgMD5  = "MD5"
	ASCIISignerAlgSHA1 = "SHA1"
)

// ASCIISigner
type ASCIISigner struct {
	Value        interface{}
	Alg          string
	OmitKeys     []string
	IncludeEmpty bool
	PrintRaw     bool
	FieldString  func(string, interface{}, bool) string
	Prefix       string
	Suffix       string
	ToUpper      bool
}

// Sign
func (s *ASCIISigner) Sign(value ...interface{}) (signature string) {
	if s == nil {
		return
	}
	if len(value) > 0 && value[0] != nil {
		s.Value = value[0]
	}
	if s.Value == nil {
		return
	}

	if len(s.Alg) == 0 {
		s.Alg = "MD5"
	}

	// 参数解析
	params := make(map[string]interface{})
	if v, ok := s.Value.(map[string]interface{}); ok {
		params = v
	} else {
		b, _ := json.Marshal(s.Value)
		_ = json.Unmarshal(b, &params)
	}

	if len(params) == 0 {
		return
	}

	if s.FieldString == nil {
		s.FieldString = func(key string, value interface{}, last bool) string {
			if last {
				return fmt.Sprintf("%s=%v", key, value)
			} else {
				return fmt.Sprintf("%s=%v&", key, value)
			}
		}
	}

	// ASCII码排序
	keys := ASCIISort(params, s.IncludeEmpty, s.OmitKeys)

	// 拼接参数
	var buffer bytes.Buffer
	if len(s.Prefix) > 0 {
		buffer.WriteString(s.Prefix)
	}
	for index, key := range keys {
		buffer.WriteString(s.FieldString(key, params[key], (len(keys)-1) == index))
	}
	if len(s.Suffix) > 0 {
		buffer.WriteString(s.Suffix)
	}
	raw := buffer.String()
	if s.PrintRaw {
		INFO(raw)
	}
	s.Alg = strings.ToUpper(s.Alg)
	switch s.Alg {
	case ASCIISignerAlgMD5:
		signature = MD5(raw, s.ToUpper)
	case ASCIISignerAlgSHA1, "SHA-1":
		signature = Sha1(raw, s.ToUpper)
	}
	return
}

// ASCIISort
func ASCIISort(data interface{}, includeEmpty bool, omitKeys ...[]string) (sortedKeys []string) {
	params := make(map[string]interface{})

	if v, ok := data.(map[string]interface{}); ok {
		params = v
	} else {
		b, _ := json.Marshal(data)
		_ = json.Unmarshal(b, &params)
	}

	omitKeyMap := make(map[string]bool)
	if len(omitKeys) > 0 {
		for _, omitKey := range omitKeys[0] {
			omitKeyMap[omitKey] = true
		}
	}

	for k, v := range params {
		if omitKeyMap[k] {
			continue
		}
		if !includeEmpty {
			if vv, ok := v.(string); ok && vv == "" {
				continue
			}
		}
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return
}
