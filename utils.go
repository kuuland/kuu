package kuu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Join 基于字节实现的字符串拼接函数
func Join(args ...string) string {
	b := bytes.Buffer{}
	for _, item := range args {
		b.WriteString(item)
	}
	return b.String()
}

// CopyBody 复制请求体
func CopyBody(c *gin.Context, docs interface{}) (err error) {
	var buf bytes.Buffer
	dst := io.TeeReader(c.Request.Body, &buf)
	if data, err := ioutil.ReadAll(dst); err == nil {
		json.Unmarshal(data, docs)
	}
	return err
}

// JSONConvert 用于对类JSON数据的快速转换
func JSONConvert(s, t interface{}) {
	if b, e := json.Marshal(s); e == nil {
		json.Unmarshal(b, t)
	}
}

// EnsureDir 确保文件夹存在
func EnsureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// EmptyDir 确保文件夹存在且为空文件夹
func EmptyDir(dir string) {
	if dir == "/" {
		Error("What are you doing?")
		return
	}
	stat, _ := os.Stat(dir)
	if stat != nil && stat.IsDir() {
		os.RemoveAll(dir)
	}
	EnsureDir(dir)
}

// ToInterfaceArray 将传入值转换为[]interface{}类型
func ToInterfaceArray(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return make([]interface{}, 0, 0)
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret
}

// GetKind 通过反射获取传入值的Kind信息
func GetKind(a interface{}) reflect.Kind {
	v := reflect.ValueOf(a)
	kind := v.Kind()
	if kind == reflect.Ptr {
		v = v.Elem()
		kind = v.Kind()
	}
	return kind
}

// IsPtr 检查是否指针
func IsPtr(a interface{}) bool {
	v := reflect.ValueOf(a)
	return v.Kind() == reflect.Ptr
}

// IsArray 判断传入值是否为数组
func IsArray(a interface{}) bool {
	kind := GetKind(a)
	return kind == reflect.Slice || kind == reflect.Array
}

// GoID 获取Goroutine ID
func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	s := string(buf[:n])
	idField := strings.Fields(strings.TrimPrefix(s, "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// SetGoroutineCache 设置线程缓存
func SetGoroutineCache(key string, val interface{}) {
	goid := strconv.Itoa(GoID())
	var data H
	if v, ok := Data[goid].(H); ok {
		data = v
	} else {
		data = make(H)
	}
	data[key] = val
	Data[goid] = data
}

// GetGoroutineCache 获取线程缓存中的指定键的值
func GetGoroutineCache() H {
	goid := strconv.Itoa(GoID())
	if v, ok := Data[goid].(H); ok {
		return v
	}
	return H{}
}

// DelGoroutineCache 删除线程缓存中的指定键
func DelGoroutineCache(keys ...string) {
	goid := strconv.Itoa(GoID())
	var data H
	if v, ok := Data[goid].(H); ok {
		data = v
	} else {
		data = make(H)
	}
	for _, key := range keys {
		delete(data, key)
	}
	Data[goid] = data
}

// ClearGoroutineCache 清空线程所有缓存
func ClearGoroutineCache() {
	goid := strconv.Itoa(GoID())
	delete(Data, goid)
}

// Cwd 当前程序运行目录
func Cwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
