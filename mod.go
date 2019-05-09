package kuu

import (
	"github.com/gin-gonic/gin"
	"path"
)

// Mod
type Mod struct {
	Routes     Routes
	Middleware Middleware
	Models     []interface{}
}

// RouteInfo
type RouteInfo struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// Routes
type Routes []RouteInfo

// Middleware
type Middleware []gin.HandlerFunc

// Import
func Import(r *gin.Engine, mods ...*Mod) {
	for _, mod := range mods {
		for _, middleware := range mod.Middleware {
			if middleware != nil {
				r.Use(middleware)
			}
		}
		for _, route := range mod.Routes {
			if route.Path == "" || route.Handler == nil {
				continue
			}
			if route.Method == "" {
				route.Method = "GET"
			}
			r.Handle(route.Method, path.Join(C().GetString("prefix"), route.Path), route.Handler)
		}
		for _, model := range mod.Models {
			if model == nil {
				continue
			}
			// TODO 挂载模型接口
		}
	}
}
