package kuu

import (
	"github.com/gin-gonic/gin"
)

// KuuMethods 全局API集合
type KuuMethods map[string]func(...interface{}) interface{}

// AppMethods 实例API集合
type AppMethods map[string]func(*Kuu, ...interface{}) interface{}

// Plugin 插件
type Plugin struct {
	Name       string
	Routes     Routes
	Middleware Middleware
	KuuMethods KuuMethods
	AppMethods AppMethods
}

// RouteInfo 路由声明
type RouteInfo struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// Routes 路由数组
type Routes []RouteInfo

// Middleware 中间件数组
type Middleware []gin.HandlerFunc

// D 调用插件API
func D(name string, args ...interface{}) interface{} {
	fn := kuuMethods[name]
	if fn == nil {
		return nil
	}
	return fn(args...)
}

// DString 返回字符串
func DString(name string, args ...interface{}) string {
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
	fn := k.appMethods[name]
	if fn == nil {
		return nil
	}
	return fn(k, args...)
}

// DString 返回字符串
func (k *Kuu) DString(name string, args ...interface{}) string {
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
