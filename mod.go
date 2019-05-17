package kuu

import (
	"github.com/gin-gonic/gin"
	"path"
	"reflect"
)

var metadata map[string]*Metadata

// Mod
type Mod struct {
	Middleware  gin.HandlersChain
	Routes      gin.RoutesInfo
	Models      []interface{}
	AfterImport func()
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
		var metaArr []*Metadata
		for _, model := range mod.Models {
			if model == nil {
				continue
			}

			if meta := parseMetadata(model); meta != nil {
				metaArr = append(metaArr, meta)
			}

			RESTful(r, model)
		}
		if mod.AfterImport != nil {
			mod.AfterImport()
		}
		if len(metaArr) > 0 {
			tx := DB().Begin()
			tx = tx.Unscoped().Delete(&Metadata{IsBuiltIn: true})
			for _, meta := range metaArr {
				tx = tx.Create(meta)
			}
			if errs := tx.GetErrors(); len(errs) > 0 {
				ERROR(errs)
				if err := tx.Rollback(); err != nil {
					ERROR(err)
				}
			} else {
				if err := tx.Commit(); err != nil {
					ERROR(err)
				} else {
					for _, meta := range metaArr {
						metadata[meta.Name] = meta
					}
				}
			}
		}
	}
}

func parseMetadata(value interface{}) (m *Metadata) {
	if value == nil {
		return
	}
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
func Meta(name string) (m *Metadata) {
	return metadata[name]
}

// Metalist
func Metalist() (arr []*Metadata) {
	for _, v := range metadata {
		arr = append(arr, v)
	}
	return
}
