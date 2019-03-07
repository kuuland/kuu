package kuu

import (
	"github.com/gin-gonic/gin"
)

// Mod 模块实例
type Mod struct {
	Routes     Routes
	Middleware Middleware
	Models     []interface{}
	Langs      map[string]LangMessages
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
