package kuu

import (
	"github.com/gin-gonic/gin"
	"path"
	"reflect"
)

var metadata = make(map[string]*Metadata)

// Mod
type Mod struct {
	Middleware  gin.HandlersChain
	Routes      KuuRoutesInfo
	Models      []interface{}
	AfterImport func()
}

// Import
func (e *Engine) Import(mods ...*Mod) {
	for _, mod := range mods {
		for _, middleware := range mod.Middleware {
			if middleware != nil {
				e.Engine.Use(middleware)
			}
		}
	}
	for _, mod := range mods {
		for _, route := range mod.Routes {
			if route.Path == "" || route.HandlerFunc == nil {
				PANIC("Route path and handler can't be nil")
			}
			if route.Method == "" {
				route.Method = "GET"
			}
			routePath := path.Join(C().GetString("prefix"), route.Path)
			if route.Method == "*" {
				e.Any(routePath, route.HandlerFunc)
			} else {
				e.Handle(route.Method, routePath, route.HandlerFunc)
			}
		}
		for _, model := range mod.Models {
			RESTful(e, model)
			if meta := parseMetadata(model); meta != nil {
				metadata[meta.Name] = meta
			}
		}
		if mod.AfterImport != nil {
			mod.AfterImport()
		}
	}
}

func parseMetadata(value interface{}) (m *Metadata) {
	reflectType := reflect.ValueOf(value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Scope value need to be a struct
	if reflectType.Kind() != reflect.Struct {
		return
	}

	m = new(Metadata)
	m.IsBuiltIn = true
	m.Name = reflectType.Name()
	m.FullName = path.Join(reflectType.PkgPath(), m.Name)
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)
		m.Fields = append(m.Fields, MetadataField{
			Name: fieldStruct.Name,
			Type: fieldStruct.Type.Kind().String(),
		})
	}
	return
}

// Meta
func Meta(value interface{}) (m *Metadata) {
	if v, ok := value.(string); ok {
		return metadata[v]
	} else {
		return parseMetadata(value)
	}
}

// Metalist
func Metalist() (arr []*Metadata) {
	for _, v := range metadata {
		arr = append(arr, v)
	}
	return
}

// RegisterMeta
func RegisterMeta() {
	tx := DB().Begin()
	tx = tx.Unscoped().Where(&Metadata{IsBuiltIn: true}).Delete(&Metadata{})
	for _, meta := range metadata {
		tx = tx.Create(meta)
	}
	if errs := tx.GetErrors(); len(errs) > 0 {
		ERROR(errs)
		if err := tx.Rollback(); err != nil {
			ERROR(err)
		}
	} else {
		if err := tx.Commit().Error; err != nil {
			ERROR(err)
		}
	}
}
