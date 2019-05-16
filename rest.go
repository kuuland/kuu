package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var modelStructsMap sync.Map

type PreloadHooks interface {
	QueryPreload(*gorm.DB) *gorm.DB
}

func callPreloadHooks(db *gorm.DB, value interface{}) *gorm.DB {
	if h, ok := value.(PreloadHooks); ok {
		db = h.QueryPreload(db)
	}
	return db
}

// RESTful
func RESTful(r *gin.Engine, value interface{}) {
	// Scope value can't be nil
	if value == nil {
		return
	}

	DB().AutoMigrate(value)
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
		if strings.Contains(string(fieldStruct.Tag), "rest") {
			// mounted RESTful routes
			tagSettings := parseTagSetting(fieldStruct.Tag, "rest")
			routeAlias := strings.TrimSpace(fieldStruct.Tag.Get("route"))
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

			if _, exists := tagSettings["-"]; exists {
				createMethod = "-"
				deleteMethod = "-"
				queryMethod = "-"
				updateMethod = "-"
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
							STDErr(c, "Parsing body failed")
							return
						}
						indirectScopeValue := indirect(reflect.ValueOf(body))
						tx := DB().Begin()
						var (
							docs  []interface{}
							errs  []error
							multi bool
						)
						if indirectScopeValue.Kind() == reflect.Slice {
							multi = true
							for i := 0; i < indirectScopeValue.Len(); i++ {
								doc := reflect.New(reflectType).Interface()
								GetSoul(indirectScopeValue.Index(i).Interface(), doc)
								tx := tx.Create(doc)
								if es := tx.GetErrors(); len(es) > 0 {
									errs = append(errs, es...)
								} else {
									docs = append(docs, doc)
								}
							}
						} else {
							doc := reflect.New(reflectType).Interface()
							GetSoul(body, doc)
							tx = tx.Create(doc)
							if es := tx.GetErrors(); len(es) > 0 {
								errs = append(errs, es...)
							} else {
								docs = append(docs, doc)
							}
						}
						if errs = txerrs(tx, errs); len(errs) > 0 {
							msg := L(c, "rest_create_failed", "Create %s failed", structName)
							ERROR(msg)
							ERROR(errs)
							STDErr(c, msg)
						} else {
							if multi {
								STD(c, docs)
							} else {
								STD(c, docs[0])
							}
						}
					})
				}
				if deleteMethod != "-" {
					r.Handle(deleteMethod, routePath, func(c *gin.Context) {
						var params map[string]interface{}
						if err := c.ShouldBindBodyWith(&params, binding.JSON); err != nil {
							ERROR(err)
							STDErr(c, "Parsing body failed")
							return
						}
						if params == nil || IsBlank(params) {
							STDErr(c, "'cond' is required")
							return
						}
						_, multi := params["multi"]
						if IsBlank(params["cond"]) && !multi {
							STDErr(c, "'multi' is required for batch delete")
							return
						}
						tx := DB().Begin()
						var (
							value interface{}
							errs  []error
						)
						if multi {
							value = reflect.New(reflect.SliceOf(reflectType)).Interface()
							tx = tx.Where(params["cond"])
							tx = callPreloadHooks(tx, reflect.New(reflectType).Interface())
							tx = tx.Find(value).Delete(reflect.New(reflectType).Interface())
							errs = tx.GetErrors()
						} else {
							value = reflect.New(reflectType).Interface()
							tx = tx.Where(params["cond"])
							tx = callPreloadHooks(tx, value)
							tx = tx.First(value).Delete(value)
							errs = tx.GetErrors()
						}
						if errs = txerrs(tx, errs); len(errs) > 0 {
							msg := L(c, "rest_delete_failed", "Delete %s failed", structName)
							ERROR(msg)
							ERROR(errs)
							STDErr(c, msg)
						} else {
							STD(c, value)
						}
					})
				}
				if queryMethod != "-" {
					r.Handle(queryMethod, routePath, func(c *gin.Context) {
						ret := map[string]interface{}{}
						// 处理cond
						var cond map[string]interface{}
						rawCond := c.Query("cond")
						if rawCond != "" {
							Parse(rawCond, &cond)
							ret["cond"] = cond
						}
						db := DB().Model(reflect.New(reflectType).Interface())
						if cond != nil {
							for key, val := range cond {
								if key == "$and" || key == "$or" {
									// arr = [{"pass":"123"},{"pass":{"$regex":"^333"}}]
									if arr, ok := val.([]interface{}); ok {
										queries := make([]string, 0)
										args := make([]interface{}, 0)
										for _, item := range arr {
											// obj = {"pass":"123"}
											obj, ok := item.(map[string]interface{})
											if !ok {
												continue
											}
											for k, v := range obj {
												if m, ok := v.(map[string]interface{}); ok {
													q, a := fieldQuery(m, k)
													if !IsBlank(q) && !IsBlank(a) {
														queries = append(queries, q)
														args = append(args, a...)
													}
												} else {
													queries = append(queries, fmt.Sprintf("\"%s\" = ?", k))
													args = append(args, v)
												}
											}
										}
										if !IsBlank(queries) && !IsBlank(args) {
											if key == "$or" {
												db = db.Where(strings.Join(queries, " OR "), args...)
											} else {
												db = db.Where(strings.Join(queries, " AND "), args...)
											}
										}
									}
									delete(cond, key)
								} else if m, ok := val.(map[string]interface{}); ok {
									query, args := fieldQuery(m, key)
									db = db.Where(query, args...)
									delete(cond, key)
								}
							}
							if !IsBlank(cond) {
								db = db.Where(cond)
							}
						}
						countDB := db
						// 处理project
						rawProject := c.Query("project")
						if rawProject != "" {
							fields := strings.Split(rawProject, ",")
							for index, field := range fields {
								if strings.HasPrefix(field, "-") {
									fields[index] = field[1:]
								}
							}
							db = db.Select(fields)
							ret["project"] = rawProject
						}
						// 处理sort
						rawSort := c.Query("sort")
						if rawSort != "" {
							split := strings.Split(rawSort, ",")
							for _, name := range split {
								if strings.HasPrefix(name, "-") {
									db = db.Order(fmt.Sprintf("%s desc", name[1:]))
								} else {
									db = db.Order(name)
								}
							}
							ret["sort"] = rawSort
						}
						// 处理range
						rawRange := strings.ToUpper(c.DefaultQuery("range", "PAGE"))
						ret["range"] = rawRange
						// 处理page、size
						var (
							page int
							size int
						)
						rawPage := c.DefaultQuery("page", "1")
						rawSize := c.DefaultQuery("size", "30")
						if v, err := strconv.Atoi(rawPage); err == nil {
							page = v
						}
						if v, err := strconv.Atoi(rawSize); err == nil {
							size = v
						}
						if rawRange == "PAGE" {
							db = db.Offset((page - 1) * size).Limit(size)
							ret["page"] = page
							ret["size"] = size
						}

						list := reflect.New(reflect.SliceOf(reflectType)).Interface()
						db = callPreloadHooks(db, reflect.New(reflectType).Interface())
						db = db.Find(list)
						ret["list"] = list
						// 处理totalrecords、totalpages
						var totalRecords int
						countDB = countDB.Count(&totalRecords)
						ret["totalrecords"] = totalRecords
						if rawRange == "PAGE" {
							ret["totalpages"] = int(math.Ceil(float64(totalRecords) / float64(size)))
						}
						STD(c, ret)
					})
				}
				if updateMethod != "-" {
					r.Handle(updateMethod, routePath, func(c *gin.Context) {
						var params map[string]interface{}
						if err := c.ShouldBindBodyWith(&params, binding.JSON); err != nil {
							ERROR(err)
							STDErr(c, "Parsing body failed")
							return
						}
						if params == nil || IsBlank(params) {
							STDErr(c, "'cond' and 'doc' are required")
							return
						}
						_, multi := params["multi"]
						if IsBlank(params["cond"]) && !multi {
							STDErr(c, "'multi' is required for batch update")
							return
						}
						if IsBlank(params["doc"]) {
							STDErr(c, "'doc' is required")
							return
						}

						// 执行更新
						tx := DB().Begin()
						msg := L(c, "rest_update_failed", "Update %s failed", structName)
						var errs []error
						if multi {
							tx = tx.Model(reflect.New(reflectType).Interface()).Where(params["cond"]).Updates(params["doc"])
							if errs = txerrs(tx, tx.GetErrors()); len(errs) > 0 {
								ERROR(msg)
								ERROR(errs)
								STDErr(c, msg)
								return
							}
						} else {
							cond := reflect.New(reflectType).Interface()
							tx = callPreloadHooks(tx, cond)
							tx = tx.First(cond).Model(cond).Updates(params["doc"])
							if errs = txerrs(tx, tx.GetErrors()); len(errs) > 0 {
								ERROR(msg)
								ERROR(errs)
								STDErr(c, msg)
								return
							}
						}
						// 查询更新后的数据
						var data interface{}
						value := reflect.New(reflectType).Interface()
						query := DB().Model(value).Where(params["cond"])
						query = callPreloadHooks(query, value)
						if multi {
							query = query.Find(data)
						} else {
							query = query.First(data)
						}
						STD(c, data)
					})
				}
			}
			break
		}
	}
}

func txerrs(tx *gorm.DB, errs []error) []error {
	if len(errs) > 0 {
		if err := tx.Rollback().Error; err != nil {
			errs = append(errs, err)
		}
	} else {
		if err := tx.Commit().Error; err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func fieldQuery(m map[string]interface{}, key string) (query string, args []interface{}) {
	if raw, has := m["$regex"]; has {
		keyword := raw.(string)
		hasPrefix := strings.HasPrefix(keyword, "^")
		hasSuffix := strings.HasSuffix(keyword, "$")
		if hasPrefix {
			keyword = keyword[1:]
		}
		if hasSuffix {
			keyword = keyword[:len(keyword)-1]
		}
		a := make([]string, 0)
		if hasPrefix {
			a = append(a, "%")
		}
		a = append(a, keyword)
		if hasSuffix {
			a = append(a, "%")
		}
		keyword = strings.Join(a, "")
		return fmt.Sprintf("\"%s\" LIKE ?", key), []interface{}{keyword}
	} else if raw, has := m["$in"]; has {
		return fmt.Sprintf("\"%s\" IN (?)", key), []interface{}{raw}
	} else if raw, has := m["$nin"]; has {
		return fmt.Sprintf("\"%s\" NOT IN (?)", key), []interface{}{raw}
	} else if raw, has := m["$eq"]; has {
		return fmt.Sprintf("\"%s\" = ?", key), []interface{}{raw}
	} else if raw, has := m["$ne"]; has {
		return fmt.Sprintf("\"%s\" <> ?", key), []interface{}{raw}
	} else if raw, has := m["$exists"]; has {
		return fmt.Sprintf("\"%s\" IS NOT NULL", key), []interface{}{raw}
	} else {
		gt, hgt := m["$gt"]
		gte, hgte := m["$gte"]
		lt, hlt := m["$lt"]
		lte, hlte := m["$lte"]
		if hgt {
			if hlt {
				return fmt.Sprintf("\"%s\" > ? AND \"%s\" < ?", key, key), []interface{}{gt, lt}
			} else if hlte {
				return fmt.Sprintf("\"%s\" > ? AND \"%s\" <= ?", key, key), []interface{}{gt, lte}
			} else {
				return fmt.Sprintf("\"%s\" > ?", key), []interface{}{gt}
			}
		} else if hgte {
			if hlt {
				return fmt.Sprintf("\"%s\" >= ? AND \"%s\" < ?", key, key), []interface{}{gte, lt}
			} else if hlte {
				return fmt.Sprintf("\"%s\" >= ? AND \"%s\" <= ?", key, key), []interface{}{gte, lte}
			} else {
				return fmt.Sprintf("\"%s\" >= ?", key), []interface{}{gte}
			}
		} else if hlt {
			return fmt.Sprintf("\"%s\" < ?", key), []interface{}{lt}
		} else if hlte {
			return fmt.Sprintf("\"%s\" <= ?", key), []interface{}{lte}
		}
	}
	return
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
