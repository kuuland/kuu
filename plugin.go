package kuu

import (
	"github.com/gin-gonic/gin"
)

// R 路由集合
type R map[string]*Route

// M 中间件集合
type M map[string]gin.HandlerFunc

// Method 全局API集合
type Method map[string]func(...interface{}) interface{}

// InstMethod 实例API集合
type InstMethod map[string]func(*Kuu, ...interface{}) interface{}

// Plugin 插件
type Plugin struct {
	Name        string
	Routes      R
	Middleware  M
	Methods     Method
	InstMethods InstMethod
	Onload      func(k *Kuu)
}

// Route 路由描述体
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}
