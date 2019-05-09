package kuu

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

// IsBlank
func IsBlank(value interface{}) bool {
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
	case reflect.Interface, reflect.Ptr:
		return indirectValue.IsNil()
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

// Parse
func Parse(v string, r interface{}) {
	err := json.Unmarshal([]byte(v), r)
	if err != nil {
		ERROR(err)
	}
}
