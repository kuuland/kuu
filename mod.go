package kuu

import (
	"fmt"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
)

var (
	tableNames       = make(map[string]string)
	tableNameMetaMap = make(map[string]*Metadata)
	ModMap           = make(map[string]*Mod)
)

func init() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		v, ok := tableNames[defaultTableName]
		if !ok || v == "" {
			if defaultTableName != "" {
				WARN("自定义表名：%s", defaultTableName)
			}
			return defaultTableName
		}
		return v
	}
}

// Mod
type Mod struct {
	Code        string
	Prefix      string
	Middleware  gin.HandlersChain
	Routes      RoutesInfo
	Models      []interface{}
	AfterImport func()
}

// Import
func (e *Engine) Import(mods ...*Mod) {
	if err := CatchError(func() {
		migrate := C().GetBool("gorm:migrate")
		for _, mod := range mods {
			for _, middleware := range mod.Middleware {
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
			routePrefix := path.Join(C().GetString("prefix"), mod.Prefix)
			for _, route := range mod.Routes {
				if route.Path == "" || route.HandlerFunc == nil {
					PANIC("Route path and handler can't be nil")
				}
				if route.Name == "" {
					PANIC("Need route name for system audit: %s %s", route.Method, route.Path)
				}
				if route.Method == "" {
					route.Method = "GET"
				}
				var routePath string
				if route.IgnorePrefix {
					routePath = path.Join(route.Path)
				} else {
					routePath = path.Join(routePrefix, route.Path)
				}
				if route.Method == "*" {
					e.Any(routePath, route.HandlerFunc)
				} else {
					e.Handle(route.Method, routePath, route.HandlerFunc)
				}
			}
			for _, model := range mod.Models {
				desc := RESTful(e, routePrefix, model)
				if meta := parseMetadata(model); meta != nil {
					meta.RestDesc = desc
					meta.ModCode = mod.Code
					defaultTableName := gorm.ToTableName(meta.Name)
					pluralTableName := inflection.Plural(defaultTableName)

					tableName := fmt.Sprintf("%s_%s", mod.Code, meta.Name)
					tableNames[defaultTableName] = tableName
					tableNames[pluralTableName] = tableName

					db := DB()
					tn := db.NewScope(model).GetModelStruct().TableName(db)
					if tn != "" {
						tableNameMetaMap[tn] = meta
					}

				}
				if migrate {
					DB().AutoMigrate(model)

					if v, ok := model.(DBTypeRepairer); ok {
						v.RepairDBTypes()
					}
				}
			}
			if mod.AfterImport != nil {
				mod.AfterImport()
			}
			ModMap[mod.Code] = mod
		}
	}); err != nil {
		panic(err)
	}
}

// GetModPrefix
func GetModPrefix(modCode string) string {
	mod, ok := ModMap[modCode]
	if ok {
		return mod.Prefix
	}
	return ""
}
