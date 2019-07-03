package kuu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

// IsBlank
func IsBlank(value interface{}) bool {
	if value == nil {
		return true
	}
	indirectValue := indirectValue(value)
	switch indirectValue.Kind() {
	case reflect.String:
		return indirectValue.Len() == 0
	case reflect.Bool:
		return !indirectValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return indirectValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return indirectValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return indirectValue.Float() == 0
	case reflect.Map:
		return indirectValue.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return indirectValue.IsNil()
	}
	vi := reflect.ValueOf(value)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return reflect.DeepEqual(indirectValue.Interface(), reflect.Zero(indirectValue.Type()).Interface())
}

func indirectValue(value interface{}) reflect.Value {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

// CORSMiddleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Capitalize
func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				fmt.Println("Not begins with lowercase letter,")
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

// Stringify
func Stringify(v interface{}, format ...bool) (ret string) {
	if b, err := json.Marshal(v); err == nil {
		if len(format) > 0 {
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

// EnsureDir
func EnsureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			ERROR(err)
		}
	}
}

// ParseID
func ParseID(id string) uint {
	if v, err := strconv.ParseUint(id, 10, 0); err != nil {
		ERROR(err)
	} else {
		return uint(v)
	}
	return 0
}

// Parse
func Parse(v string, r interface{}) {
	err := json.Unmarshal([]byte(v), r)
	if err != nil {
		ERROR(err)
	}
}

// Copy
func Copy(src interface{}, dest interface{}) (err error) {
	var data []byte
	if data, err = json.Marshal(src); err == nil {
		err = json.Unmarshal(data, dest)
	}
	if err != nil {
		ERROR(err)
	}
	return
}


