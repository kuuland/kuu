package kuu

import (
	"github.com/gin-gonic/gin"
)

// Plugin 插件
type Plugin struct {
	Routes     Routes
	Middleware Middleware
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
