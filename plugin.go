package kuu

import (
	"github.com/gin-gonic/gin"
)

// R 路由集合
type R map[string]*Route

// M 中间件集合
type M map[string]gin.HandlerFunc

// Methods 全局API集合
type Methods map[string]func(...interface{}) interface{}

// InstMethods 实例API集合
type InstMethods map[string]func(*Kuu, ...interface{}) interface{}

// Plugin 插件
type Plugin struct {
	Name        string
	Routes      R
	Middleware  M
	Methods     Methods
	InstMethods InstMethods
	OnLoad      func(k *Kuu)
	OnModel     func(k *Kuu, schema *Schema)
	BeforeRun   func(k *Kuu)
}

// Route 路由描述体
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// D 调用插件API
func D(name string, args ...interface{}) interface{} {
	fn := methods[name]
	if fn == nil {
		return nil
	}
	return fn(args...)
}

// DStr 返回字符串
func DStr(name string, args ...interface{}) string {
	return D(name, args...).(string)
}

// DInt 返回整型
func DInt(name string, args ...interface{}) int {
	return D(name, args...).(int)
}

// DBool 返回布尔值
func DBool(name string, args ...interface{}) bool {
	return D(name, args...).(bool)
}

// DFloat 返回浮点值
func DFloat(name string, args ...interface{}) float64 {
	return D(name, args...).(float64)
}

// D 调用插件实例API
func (k *Kuu) D(name string, args ...interface{}) interface{} {
	fn := k.methods[name]
	if fn == nil {
		return nil
	}
	return fn(k, args...)
}

// DStr 返回字符串
func (k *Kuu) DStr(name string, args ...interface{}) string {
	return k.D(name, args...).(string)
}

// DInt 返回整型
func (k *Kuu) DInt(name string, args ...interface{}) int {
	return k.D(name, args...).(int)
}

// DBool 返回布尔值
func (k *Kuu) DBool(name string, args ...interface{}) bool {
	return k.D(name, args...).(bool)
}

// DFloat 返回浮点值
func (k *Kuu) DFloat(name string, args ...interface{}) float64 {
	return k.D(name, args...).(float64)
}
