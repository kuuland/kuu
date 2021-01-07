package kuu

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"gopkg.in/guregu/null.v3"
	"path"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
)

var (
	tableNames       = make(map[string]string)
	tableNameMetaMap = make(map[string]*Metadata)
	modMap           = make(map[string]*Mod)
)

var (
	routesMap   = make(map[string]RouteInfo)
	routesMapMu sync.RWMutex
)

func init() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if v, ok := tableNames[defaultTableName]; ok && v != "" {
			return v
		}
		return defaultTableName
	}
}

type tabler interface {
	TableName() string
}

type dbTabler interface {
	TableName(*gorm.DB) string
}

// HandlerFunc defines the handler used by ok middleware as return value.
type HandlerFunc func(*Context) *STDReply

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

// Mod
type Mod struct {
	Code         string
	Prefix       string
	Middleware   HandlersChain
	Routes       RoutesInfo
	Models       []interface{}
	IntlMessages map[string]string
	OnImport     func() error
	OnInit       func() error
	TablePrefix  null.String
}

// Import
func (app *Engine) Import(mods ...*Mod) {
	migrate := C().GetBool("gorm:migrate")
	for _, mod := range mods {
		for _, middleware := range mod.Middleware {
			if middleware != nil {
				app.Use(middleware)
			}
		}
	}
	for _, mod := range mods {
		if mod.Code == "" {
			PANIC("模块编码不能为空")
		}
		if len(mod.IntlMessages) > 0 {
			AddDefaultIntlMessage(mod.IntlMessages)
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
			route.Method = strings.ToUpper(route.Method)
			var routePath string
			if route.IgnorePrefix {
				routePath = path.Join(route.Path)
			} else {
				routePath = path.Join(routePrefix, route.Path)
			}
			if route.Method == "*" {
				app.Any(routePath, route.HandlerFunc)
				for _, method := range []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS", "DELETE", "CONNECT", "TRACE"} {
					key := fmt.Sprintf("%s %s", method, routePath)
					routesMapMu.Lock()
					routesMap[key] = route
					routesMapMu.Unlock()
				}
			} else {
				app.Handle(route.Method, routePath, route.HandlerFunc)
				key := fmt.Sprintf("%s %s", route.Method, routePath)
				routesMapMu.Lock()
				routesMap[key] = route
				routesMapMu.Unlock()
			}
			if len(route.IntlMessages) > 0 {
				AddDefaultIntlMessage(route.IntlMessages)
			}
		}
		for _, model := range mod.Models {
			desc := RESTful(app, routePrefix, model)
			if meta := parseMetadata(model); meta != nil {
				meta.RestDesc = desc
				meta.ModCode = mod.Code
				gormTableName := gorm.ToTableName(meta.Name)
				pluralTableName := inflection.Plural(gormTableName)

				if v, ok := model.(tabler); ok {
					tableName := v.TableName()
					tableNames[tableName] = tableName
				} else if v, ok := model.(dbTabler); ok {
					tableName := v.TableName(DB())
					tableNames[tableName] = tableName
				} else if C().GetBool("legacy_table_name") {
					tableName := fmt.Sprintf("%s_%s", mod.Code, meta.Name)
					tableNames[gormTableName] = tableName
					tableNames[pluralTableName] = tableName
				} else {
					var kuuTableName string
					if v := mod.TablePrefix; v.Valid {
						if v.String != "" {
							kuuTableName = fmt.Sprintf("%s_%s", v.String, pluralTableName)
						} else {
							kuuTableName = pluralTableName
						}
					} else {
						kuuTableName = fmt.Sprintf("%s_%s", mod.Code, pluralTableName)
					}

					tableNames[gormTableName] = kuuTableName
					tableNames[pluralTableName] = kuuTableName
					tableNames[kuuTableName] = kuuTableName
				}

				metaTableName := DB().NewScope(model).TableName()
				if metaTableName != "" {
					tableNameMetaMap[metaTableName] = meta
				}
			}
			if migrate {
				DB().AutoMigrate(model)

				if v, ok := model.(DBTypeRepairer); ok {
					v.RepairDBTypes()
				}
			}
		}
		if mod.OnImport != nil {
			if err := mod.OnImport(); err != nil {
				// 模块导入失败直接退出
				panic(err)
			}
		}
		modMap[mod.Code] = mod
	}
}

// GetModPrefix
func GetModPrefix(modCode string) string {
	mod, ok := modMap[modCode]
	if ok {
		return mod.Prefix
	}
	return ""
}
