package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	metadataMap     = make(map[string]*Metadata)
	metadataList    = make([]*Metadata, 0)
	tableNames      = make(map[string]string)
	modelStructsMap sync.Map
)

func init() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		v, ok := tableNames[defaultTableName]
		if !ok || v == "" {
			PANIC("表名 %s 不存在", defaultTableName)
		}
		return v
	}
}

// Mod
type Mod struct {
	Code        string
	Middlewares gin.HandlersChain
	Routes      RoutesInfo
	Models      []interface{}
	AfterImport func()
}

// Import
func (e *Engine) Import(mods ...*Mod) {
	migrate := C().GetBool("gorm:migrate")
	for _, mod := range mods {
		for _, middleware := range mod.Middlewares {
			if middleware != nil {
				e.Engine.Use(middleware)
			}
		}
	}
	for _, mod := range mods {
		if mod.Code == "" {
			PANIC("模块编码不能为空")
		}
		mod.Code = strings.ToLower(mod.Code)
		for _, route := range mod.Routes {
			if route.Path == "" || route.HandlerFunc == nil {
				PANIC("Route path and handler can't be nil")
			}
			if route.Method == "" {
				route.Method = "GET"
			}
			var routePath string
			if route.IgnorePrefix {
				routePath = path.Join(route.Path)
			} else {
				routePath = path.Join(C().GetString("prefix"), route.Path)
			}
			if route.Method == "*" {
				e.Any(routePath, route.HandlerFunc)
			} else {
				e.Handle(route.Method, routePath, route.HandlerFunc)
			}
		}
		for _, model := range mod.Models {
			desc := RESTful(e, model)
			if meta := parseMetadata(model); meta != nil {
				meta.RestDesc = desc
				metadataMap[meta.Name] = meta
				metadataList = append(metadataList, meta)

				defaultTableName := gorm.ToTableName(meta.Name)
				pluralTableName := inflection.Plural(defaultTableName)

				tableName := fmt.Sprintf("%s_%s", mod.Code, meta.Name)
				tableNames[defaultTableName] = tableName
				tableNames[pluralTableName] = tableName
				meta.TableName = tableName

				if methodValue := indirectValue(model).MethodByName("TableName"); methodValue.IsValid() {
					var modelTableName string
					switch method := methodValue.Interface().(type) {
					case func() string:
						modelTableName = method()
					case func(*gorm.DB) string:
						modelTableName = method(DB())
					}
					if modelTableName != "" {
						meta.TableName = modelTableName
						tableNames[modelTableName] = modelTableName
					}
				}
			}
			if migrate {
				DB().AutoMigrate(model)
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

	hashKey := reflectType
	if value, ok := modelStructsMap.Load(hashKey); ok && value != nil {
		return value.(*Metadata)
	}

	m = new(Metadata)
	m.Name = reflectType.Name()
	m.FullName = path.Join(reflectType.PkgPath(), m.Name)
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)
		displayName := fieldStruct.Tag.Get("displayName")
		if m.DisplayName == "" && displayName != "" {
			m.DisplayName = displayName
		}
		indirectType := fieldStruct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}
		fieldValue := reflect.New(indirectType).Interface()
		field := MetadataField{
			Code: fieldStruct.Name,
			Kind: fieldStruct.Type.Kind().String(),
			Enum: fieldStruct.Tag.Get("enum"),
		}
		switch field.Kind {
		case "bool":
			field.Type = "boolean"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			field.Type = "integer"
		case "float32", "float64":
			field.Type = "number"
		case "slice", "struct", "ptr":
			field.Type = "object"
		default:
			field.Type = field.Kind
		}
		if _, ok := fieldValue.(*time.Time); ok {
			field.Type = "string"
		}
		ref := fieldStruct.Tag.Get("ref")
		if ref != "" {
			fieldMeta := Meta(ref)
			if fieldMeta != nil {
				field.Type = fieldMeta.Name
				field.IsRef = true
				field.Value = fieldValue
				if indirectType.Kind() == reflect.Slice {
					field.IsArray = true
				}
			}
		}
		name := fieldStruct.Tag.Get("name")
		if name != "" {
			field.Name = name
		}
		if field.Name != "" {
			m.Fields = append(m.Fields, field)
		}
	}
	modelStructsMap.Store(hashKey, m)
	return
}

// Meta
func Meta(value interface{}) (m *Metadata) {
	if v, ok := value.(string); ok {
		return metadataMap[v]
	} else {
		return parseMetadata(value)
	}
}

// Metalist
func Metalist() []*Metadata {
	return metadataList
}

// RegisterMeta
func RegisterMeta() {
	tx := DB().Begin()
	tx = tx.Unscoped().Where(&Metadata{}).Delete(&Metadata{})
	for _, meta := range metadataList {
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
