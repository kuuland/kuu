package kuu

import (
	"github.com/gin-gonic/gin"
	"path"
)

// Mod
type Mod struct {
	Middleware gin.HandlersChain
	Routes     gin.RoutesInfo
	Models     []interface{}
}

// Import
func Import(r *gin.Engine, mods ...*Mod) {
	for _, mod := range mods {
		for _, middleware := range mod.Middleware {
			if middleware != nil {
				r.Use(middleware)
			}
		}
		for _, route := range mod.Routes {
			if route.Path == "" || route.HandlerFunc == nil {
				continue
			}
			if route.Method == "" {
				route.Method = "GET"
			}
			routePath := path.Join(C().GetString("prefix"), route.Path)
			if route.Method == "*" {
				r.Any(routePath, route.HandlerFunc)
			} else {
				r.Handle(route.Method, routePath, route.HandlerFunc)
			}
		}
		for _, model := range mod.Models {
			if model == nil {
				continue
			}
			RESTful(r, model)
		}
	}
}

// NewMod
func NewMod() *Mod {
	return &Mod{}
}

// AddModel
func (m *Mod) AddModel(models ...interface{}) *Mod {
	if !IsBlank(models) {
		m.Models = append(m.Models, models...)
	}
	return m
}

// AddRoute
func (m *Mod) AddRoute(path string, handler gin.HandlerFunc, methods ...string) *Mod {
	if IsBlank(path) || handler == nil {
		return m
	}
	if IsBlank(methods) {
		methods = []string{"GET"}
	}
	for _, method := range methods {
		if IsBlank(method) {
			continue
		}
		m.Routes = append(m.Routes, gin.RouteInfo{Method: method, Path: path, HandlerFunc: handler})
	}
	return m
}

// AddRouteInfo
func (m *Mod) AddRouteInfo(routes ...gin.RouteInfo) *Mod {
	if !IsBlank(routes) {
		m.Routes = append(m.Routes, routes...)
	}
	return m
}

// AddMiddleware
func (m *Mod) AddMiddleware(middleware ...gin.HandlerFunc) *Mod {
	if !IsBlank(middleware) {
		m.Middleware = append(m.Middleware, middleware...)
	}
	return m
}
