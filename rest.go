package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"reflect"
	"strings"
	"sync"
)

var modelStructsMap sync.Map

func MountRESTful(r *gin.Engine, value interface{}) {
	// Scope value can't be nil
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

	// Get Cached model struct
	if value, ok := modelStructsMap.Load(reflectType); ok && value != nil {
		return
	}
	structName := reflectType.Name()
	routePrefix := C().GetString("prefix")
	routePath := fmt.Sprintf("%s/%s", routePrefix, strings.ToLower(structName))

	// Get all fields
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)
		if strings.ToUpper(fieldStruct.Name) == "KUU" {
			// mounted RESTful routes
			tagSettings := parseTagSetting(fieldStruct.Tag, "rest")
			routeAlias := strings.TrimSpace(fieldStruct.Tag.Get("route"))
			tableName := strings.TrimSpace(fieldStruct.Tag.Get("table"))
			if routeAlias != "" {
				routePath = fmt.Sprintf("%s/%s", routePrefix, strings.ToLower(routeAlias))
			}
			var (
				createMethod = "POST"
				deleteMethod = "DELETE"
				queryMethod  = "GET"
				updateMethod = "PUT"
			)
			for key, val := range tagSettings {
				key = strings.TrimSpace(strings.ToUpper(key))
				val = strings.TrimSpace(strings.ToUpper(val))
				switch key {
				case "C", "CREATE":
					createMethod = val
				case "D", "DELETE", "REMOVE":
					deleteMethod = val
				case "R", "READ", "QUERY", "FIND":
					queryMethod = val
				case "U", "UPDATE":
					updateMethod = val
				}
			}

			// Method conflict
			if methodConflict([]string{createMethod, deleteMethod, queryMethod, updateMethod}) {
				ERROR("restful routes method conflict:\n%s\n%s\n%s\n%s\n ",
					fmt.Sprintf(" - create %s: %-8s %s", structName, createMethod, routePath),
					fmt.Sprintf(" - delete %s: %-8s %s", structName, deleteMethod, routePath),
					fmt.Sprintf(" - update %s: %-8s %s", structName, updateMethod, routePath),
					fmt.Sprintf(" - query  %s: %-8s %s", structName, queryMethod, routePath),
				)
			} else {
				if createMethod != "-" {
					r.Handle(createMethod, routePath, func(c *gin.Context) {
						var body interface{}
						if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
							ERROR(err)
							STDErr(c, "Parsing the request body failed")
							return
						}
						indirectScopeValue := indirect(reflect.ValueOf(body))
						if indirectScopeValue.Kind() == reflect.Slice {
							arr := make([]interface{}, 0)
							for i := 0; i < indirectScopeValue.Len(); i++ {
								doc := reflect.New(reflectType).Interface()
								GetSoul(indirectScopeValue.Index(i).Interface(), doc)
								DB().Create(doc)
								arr = append(arr, doc)
							}
							STD(c, arr)
						} else {
							doc := reflect.New(reflectType).Interface()
							if err := c.ShouldBindBodyWith(doc, binding.JSON); err != nil {
								ERROR(err)
								STDErr(c, "Parsing the request body failed")
								return
							}
							DB().Create(doc)
							STD(c, doc)
						}
					})
				}
				if deleteMethod != "-" {
					r.Handle(deleteMethod, routePath, func(c *gin.Context) {
						// TODO 删除接口
						STD(c, "删除 "+tableName)
					})
				}
				if queryMethod != "-" {
					r.Handle(queryMethod, routePath, func(c *gin.Context) {
						// TODO 查询接口
						STD(c, "查询 "+tableName)
					})
				}
				if updateMethod != "-" {
					r.Handle(updateMethod, routePath, func(c *gin.Context) {
						// TODO 修改接口
						STD(c, "修改 "+tableName)
					})
				}
			}
		}
	}
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func methodConflict(arr []string) bool {
	for i, s := range arr {
		if s == "-" {
			continue
		}
		for j, t := range arr {
			if t == "-" || j == i {
				continue
			}
			if s == t {
				return true
			}
		}
	}
	return false
}

func parseTagSetting(tags reflect.StructTag, tagKey string) map[string]string {
	setting := map[string]string{}
	str := tags.Get(tagKey)
	split := strings.Split(str, ";")
	for _, value := range split {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) >= 2 {
			setting[k] = strings.Join(v[1:], ":")
		} else {
			setting[k] = k
		}
	}
	return setting
}
