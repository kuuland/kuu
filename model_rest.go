package kuu

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
)

// RestDesc
type RestDesc struct {
	Create bool
	Delete bool
	Query  bool
	Update bool
	Import bool
}

// IsValid
func (r *RestDesc) IsValid() bool {
	return r.Create || r.Delete || r.Query || r.Update
}

// RESTful
func RESTful(r *Engine, routePrefix string, value interface{}) (desc *RestDesc) {
	desc = new(RestDesc)
	// Scope value can't be nil
	if value == nil {
		PANIC("Model can't be nil")
	}
	reflectType := reflect.ValueOf(value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Scope value need to be a struct
	if reflectType.Kind() != reflect.Struct {
		PANIC("Model need to be a struct")
	}

	structName := reflectType.Name()
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
					desc.Create = true
					r.Handle(createMethod, routePath, restCreateHandler(reflectType))
				}
				if deleteMethod != "-" {
					desc.Delete = true
					r.Handle(deleteMethod, routePath, restDeleteHandler(reflectType))
				}
				if queryMethod != "-" {
					desc.Query = true
					r.Handle(queryMethod, routePath, restQueryHandler(reflectType))
				}
				if updateMethod != "-" {
					desc.Update = true
					r.Handle(updateMethod, routePath, restUpdateHandler(reflectType))
				}
			}
			break
		}
	}
	return
}

// CondDesc
type CondDesc struct {
	AndSQLs  []string
	AndAttrs []interface{}
	OrSQLs   []string
	OrAttrs  []interface{}
}

func parseSlimField(name string, value interface{}) (sqls []string, attrs []interface{}) {
	if name == "" || value == nil {
		return
	}
	if vmap, ok := value.(map[string]interface{}); ok {
		// 对象值
		if raw, has := vmap["$regex"]; has {
			keyword := raw.(string)
			hasPrefix := strings.HasPrefix(keyword, "^")
			hasSuffix := strings.HasSuffix(keyword, "$")
			if !hasPrefix && !hasSuffix {
				keyword = fmt.Sprintf("^%s$", keyword)
				hasPrefix = true
				hasSuffix = true
			}
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
			sqls = append(sqls, fmt.Sprintf(" %s LIKE ?", name))
			attrs = append(attrs, strings.Join(a, ""))
		} else if raw, has := vmap["$in"]; has {
			sqls = append(sqls, fmt.Sprintf("%s IN (?)", name))
			attrs = append(attrs, raw)
		} else if raw, has := vmap["$nin"]; has {
			sqls = append(sqls, fmt.Sprintf("%s NOT IN (?)", name))
			attrs = append(attrs, raw)
		} else if raw, has := vmap["$eq"]; has {
			sqls = append(sqls, fmt.Sprintf("%s = ?", name))
			attrs = append(attrs, raw)
		} else if raw, has := vmap["$ne"]; has {
			sqls = append(sqls, fmt.Sprintf("%s <> ?", name))
			attrs = append(attrs, raw)
		} else if raw, has := vmap["$exists"]; has {
			if v, ok := raw.(bool); ok {
				if v {
					sqls = append(sqls, fmt.Sprintf("%s IS NOT NULL", name))
				} else {
					sqls = append(sqls, fmt.Sprintf("%s IS NULL", name))
				}
			}
		} else {
			gt, hgt := vmap["$gt"]
			gte, hgte := vmap["$gte"]
			lt, hlt := vmap["$lt"]
			lte, hlte := vmap["$lte"]
			if hgt {
				if hlt {
					sqls = append(sqls, fmt.Sprintf("%s > ? AND %s < ?", name, name))
					attrs = append(attrs, gt, lt)
				} else if hlte {
					sqls = append(sqls, fmt.Sprintf("%s > ? AND %s <= ?", name, name))
					attrs = append(attrs, gt, lte)
				} else {
					sqls = append(sqls, fmt.Sprintf("%s > ?", name))
					attrs = append(attrs, gt)
				}
			} else if hgte {
				if hlt {
					sqls = append(sqls, fmt.Sprintf("%s >= ? AND %s < ?", name, name))
					attrs = append(attrs, gte, lt)
				} else if hlte {
					sqls = append(sqls, fmt.Sprintf("%s >= ? AND %s <= ?", name, name))
					attrs = append(attrs, gte, lte)
				} else {
					sqls = append(sqls, fmt.Sprintf("%s >= ?", name))
					attrs = append(attrs, gte)
				}
			} else if hlt {
				sqls = append(sqls, fmt.Sprintf("%s < ?", name))
				attrs = append(attrs, lt)
			} else if hlte {
				sqls = append(sqls, fmt.Sprintf("%s <= ?", name))
				attrs = append(attrs, lte)
			}
		}
	} else {
		// 普通值
		sqls = append(sqls, fmt.Sprintf("%s = ?", name))
		attrs = append(attrs, value)
	}
	return
}

func parseObject(filter map[string]interface{}, model interface{}) (sqls []string, attrs []interface{}) {
	if len(filter) == 0 {
		return
	}

	for key, val := range filter {
		if v, ok := val.([]interface{}); ok {
			for _, item := range v {
				if obj, ok := item.(map[string]interface{}); ok {
					ss, as := parseObject(obj, model)
					if len(ss) > 0 {
						sqls = append(sqls, ss...)
					}
					if len(as) > 0 {
						attrs = append(attrs, as...)
					}
				}
			}
		} else {
			var (
				scope           = DB().NewScope(model)
				field, hasField = scope.FieldByName(key)
				columnName      string
				ss              []string
				as              []interface{}
			)
			if hasField {
				columnName = field.DBName
			} else {
				columnName = gorm.ToColumnName(key)
			}
			if refCond, ok := val.(map[string]interface{}); ok && field.Relationship != nil {
				ss, as = parseRefField(field, refCond)
			} else {
				ss, as = parseSlimField(columnName, val)
			}
			if len(ss) > 0 {
				sqls = append(sqls, ss...)
			}
			if len(as) > 0 {
				attrs = append(attrs, as...)
			}
		}
	}

	return
}

// ParseCond parse the cond parameter.
func ParseCond(cond interface{}, model interface{}, with ...*gorm.DB) (desc CondDesc, db *gorm.DB) {
	var (
		data  map[string]interface{}
		scope = DB().NewScope(model)
	)
	switch cond.(type) {
	case string:
		_ = Parse(cond.(string), data)
	case map[string]interface{}:
		data = cond.(map[string]interface{})
	default:
		_ = Copy(cond, data)
	}

	if len(data) > 0 {
		for key, val := range data {
			if v, ok := val.([]interface{}); ok {
				// 处理顶级$and、$or
				var (
					sqls  []string
					attrs []interface{}
				)
				for _, item := range v {
					if obj, ok := item.(map[string]interface{}); ok {
						// 只处理对象值
						var (
							ss []string
							as []interface{}
						)
						ss, as = parseObject(obj, model)
						if len(ss) > 0 {
							sqls = append(sqls, ss...)
						}
						if len(as) > 0 {
							attrs = append(attrs, as...)
						}
					}
				}
				switch key {
				case "$and":
					desc.AndSQLs = append(desc.AndSQLs, sqls...)
					desc.AndAttrs = append(desc.AndAttrs, attrs...)
				case "$or":
					desc.OrSQLs = append(desc.OrSQLs, sqls...)
					desc.OrAttrs = append(desc.OrAttrs, attrs...)
				}
			} else {
				// 处理字段键值对
				var (
					field, hasField = scope.FieldByName(key)
					columnName      string
					ss              []string
					as              []interface{}
				)
				if hasField {
					columnName = field.DBName
				}
				if refCond, ok := val.(map[string]interface{}); ok && field.Relationship != nil {
					ss, as = parseRefField(field, refCond)
				} else {
					ss, as = parseSlimField(columnName, val)
				}
				if len(ss) > 0 {
					desc.AndSQLs = append(desc.AndSQLs, ss...)
				}
				if len(as) > 0 {
					desc.AndAttrs = append(desc.AndAttrs, as...)
				}
			}
		}
	}

	if len(with) > 0 && with[0] != nil {
		db = with[0]

		if len(desc.AndSQLs) > 0 {
			db = db.Where(strings.Join(desc.AndSQLs, " AND "), desc.AndAttrs...)
		}
		if len(desc.OrSQLs) > 0 {
			db = db.Where(strings.Join(desc.OrSQLs, " OR "), desc.OrAttrs...)
		}
	}
	return
}

func parseRefField(field *gorm.Field, refCond map[string]interface{}) (sqls []string, attrs []interface{}) {
	refModel := reflect.New(field.Struct.Type).Interface()
	refScope := DB().NewScope(refModel)
	ss, as := parseObject(refCond, refModel)

	switch field.Relationship.Kind {
	case "belongs_to", "has_many", "has_one":
		var (
			srcNames  []string
			destNames []string
		)
		if field.Relationship.Kind == "belongs_to" {
			srcNames = field.Relationship.ForeignDBNames
			destNames = field.Relationship.AssociationForeignDBNames
		} else {
			srcNames = field.Relationship.AssociationForeignDBNames
			destNames = field.Relationship.ForeignDBNames
		}
		if len(srcNames) > 0 && len(destNames) > 0 {
			for _, srcName := range srcNames {
				for _, destName := range destNames {
					handler := field.Relationship.JoinTableHandler
					tableName := refScope.TableName()
					if handler != nil {
						tableName = handler.Table(refScope.DB())
					}
					refDB := DB().Table(tableName).Select(destName).Where(strings.Join(ss, " AND "), as...)
					sqls = append(sqls, fmt.Sprintf("%s IN (?)", refDB.Dialect().Quote(srcName)))
					attrs = append(attrs, refDB.QueryExpr())
				}
			}
		}
	case "many_to_many":
		if handler := field.Relationship.JoinTableHandler; handler != nil {
			tableName := handler.Table(refScope.DB())

			foreignFieldName := field.Relationship.ForeignFieldNames[0]
			foreignDBName := field.Relationship.ForeignDBNames[0]
			assForeignFieldName := field.Relationship.AssociationForeignFieldNames[0]
			assForeignDBName := field.Relationship.AssociationForeignDBNames[0]

			destDB := DB().Table(refScope.TableName()).Select(assForeignFieldName).Where(strings.Join(ss, " AND "), as...)
			refDB := DB().Table(tableName).Select(foreignDBName).Where(fmt.Sprintf("%s IN (?)", destDB.Dialect().Quote(assForeignDBName)), destDB.QueryExpr())

			sqls = append(sqls, fmt.Sprintf("%s IN (?)", refDB.Dialect().Quote(foreignFieldName)))
			attrs = append(attrs, refDB.QueryExpr())
		}
	}
	return
}

func restUpdateHandler(reflectType reflect.Type) func(c *Context) {
	return func(c *Context) {
		var (
			result     interface{}
			err        error
			modelValue = reflect.New(reflectType).Elem().Addr().Interface()
		)
		// 事务执行
		err = c.WithTransaction(func(tx *gorm.DB) error {
			var params BizUpdateParams
			if err := c.ShouldBindBodyWith(&params, binding.JSON); err != nil {
				return err
			}
			if IsBlank(params.Cond) || IsBlank(params.Doc) {
				return errors.New("'cond' and 'doc' are required")
			}
			var multi bool
			if params.Multi || params.All {
				multi = true
			}
			if IsBlank(params.Cond) && !multi {
				return errors.New("'multi' is required")
			}
			// 处理更新条件
			queryDB := tx.New()
			_, queryDB = ParseCond(params.Cond, modelValue, queryDB)
			// 先查询更新前的数据
			if multi {
				result = reflect.New(reflect.SliceOf(reflectType)).Interface()
				queryDB = queryDB.Find(result)
			} else {
				result = reflect.New(reflectType).Interface()
				queryDB = queryDB.First(result)
			}
			if queryDB.RowsAffected < 1 {
				return ErrAffectedSaveToken
			}
			updateFields := func(val interface{}) error {
				doc := reflect.New(reflectType).Interface()
				if err := Copy(params.Doc, doc); err != nil {
					return err
				}
				bizScope := NewBizScope(c, val, tx)
				bizScope.UpdateParams = &params
				bizScope.UpdateCond = val
				bizScope.Value = doc
				bizScope.callCallbacks(BizUpdateKind)
				if bizScope.HasError() {
					return bizScope.DB.Error
				}
				return tx.Error
			}
			if indirectScopeValue := indirectValue(result); indirectScopeValue.Kind() == reflect.Slice {
				for i := 0; i < indirectScopeValue.Len(); i++ {
					item := indirectScopeValue.Index(i).Interface()
					if err := updateFields(item); err != nil {
						return err
					}
				}
			} else {
				if err := updateFields(result); err != nil {
					return err
				}
			}
			return tx.Error
		})
		// 响应结果
		if err != nil {
			if cusErr, ok := ErrOut(err); ok {
				c.STDErr(c.L("kuu_error_"+fmt.Sprintf("%v", cusErr.Code), ErrMsgs(err)[0]), err)
			} else {
				c.STDErr(c.L("rest_update_failed", "Update failed"), err)
			}
		} else {
			result = Meta(reflect.New(reflectType).Interface()).OmitPassword(result)
			c.STD(result)
		}
	}
}

func restQueryHandler(reflectType reflect.Type) func(c *Context) {
	return func(c *Context) {
		var (
			modelValue = reflect.New(reflectType).Elem().Addr().Interface()
			ret        = new(BizQueryResult)
			scope      = DB().NewScope(modelValue)
		)
		// 处理cond
		var cond map[string]interface{}
		rawCond := c.Query("cond")
		if rawCond != "" {
			_ = Parse(rawCond, &cond)
			var retCond map[string]interface{}
			_ = Parse(rawCond, &retCond)
			ret.Cond = retCond
		}
		_, db := ParseCond(cond, modelValue, DB().Model(modelValue))
		// 处理project
		rawProject := c.Query("project")
		if rawProject != "" {
			split := strings.Split(rawProject, ",")
			var (
				retProject []string
				columns    []string
			)
			for _, name := range split {
				if strings.HasPrefix(name, "-") {
					name = name[1:]
				}
				if field, ok := scope.FieldByName(name); ok {
					columns = append(columns, field.DBName)
					retProject = append(retProject, field.Name)
				}
			}
			db = db.Select(columns)
			ret.Project = strings.Join(retProject, ",")
		}
		// 处理sort
		rawSort := c.Query("sort")
		if rawSort != "" {
			split := strings.Split(rawSort, ",")
			var retSort []string
			for _, name := range split {
				direction := "asc"
				if strings.HasPrefix(name, "-") {
					name = name[1:]
					direction = "desc"
				}
				if field, ok := scope.FieldByName(name); ok {
					db = db.Order(fmt.Sprintf("%s %s", field.DBName, direction))
					if direction == "desc" {
						retSort = append(retSort, "-"+field.Name)
					} else {
						retSort = append(retSort, field.Name)
					}
				}
			}
			ret.Sort = strings.Join(retSort, ",")
		} else {
			if scope.HasColumn("created_at") {
				db = db.Order("created_at desc")
			}
		}
		// 处理preload
		rawPreload := c.Query("preload")
		if rawPreload != "" {
			ms := db.NewScope(reflect.New(reflectType).Interface())
			split := strings.Split(rawPreload, ",")
			handlers := make(map[string]func(*gorm.DB) *gorm.DB)
			if v, ok := ms.Value.(BizPreloadInterface); ok {
				handlers = v.BizPreloadHandlers()
			}
			for _, item := range split {
				if handler, has := handlers[item]; has {
					db = handler(db)
					continue
				}

				tmp := item
				if strings.Contains(item, ".") {
					tmp = strings.Split(item, ".")[0]
				}

				field, ok := ms.FieldByName(tmp)
				if !ok {
					continue
				}

				if field.Relationship.Kind == "many_to_many" {
					var (
						preloadCondition string
						refTableName     = field.Relationship.JoinTableHandler.Table(db)
						refMeta          = tableNameMetaMap[refTableName]
					)
					if refMeta != nil {
						refScope := db.NewScope(reflect.New(refMeta.reflectType).Interface())
						deletedAtField, hasDeletedAt := refScope.FieldByName("DeletedAt")
						if hasDeletedAt {
							preloadCondition = fmt.Sprintf("%v.%v IS NULL",
								ms.Quote(refTableName),
								ms.Quote(deletedAtField.DBName),
							)
						}
					}
					db = db.Preload(field.Name, preloadCondition)
				} else {
					db = db.Preload(item)
				}
			}
			ret.Preload = rawPreload
		}

		ret.List = reflect.New(reflect.SliceOf(reflectType)).Interface()
		// 处理range
		rawRange := strings.ToUpper(c.DefaultQuery("range", "PAGE"))
		ret.Range = rawRange
		// 处理page、size
		page, size := c.GetPagination()
		if rawRange == "PAGE" {
			db = db.Offset((page - 1) * size).Limit(size)
			ret.Page = page
			ret.Size = size
		}
		// 调用钩子
		bizScope := NewBizScope(c, reflect.New(reflectType).Elem().Addr().Interface(), db)
		bizScope.QueryResult = ret
		bizScope.callCallbacks(BizQueryKind)
		if bizScope.HasError() {
			if cusErr, ok := ErrOut(bizScope.DB.Error); ok {
				c.STDErr(c.L("kuu_error_"+fmt.Sprintf("%v", cusErr.Code), ErrMsgs(bizScope.DB.Error)[0]), bizScope.DB.Error)
			} else {
				c.STDErr(c.L("rest_query_failed", "Query failed"), bizScope.DB.Error)
			}
			return
		}
		c.STD(ret)
	}
}

func restDeleteHandler(reflectType reflect.Type) func(c *Context) {
	return func(c *Context) {
		var (
			result     interface{}
			err        error
			modelValue = reflect.New(reflectType).Elem().Addr().Interface()
		)
		// 事务执行
		err = c.WithTransaction(func(tx *gorm.DB) error {
			var params struct {
				All    bool
				Multi  bool
				UnSoft bool
				Cond   map[string]interface{}
			}
			if c.Query("cond") != "" {
				var retCond map[string]interface{}
				Parse(c.Query("cond"), &retCond)
				params.Cond = retCond

				if c.Query("multi") != "" || c.Query("all") != "" {
					params.Multi = true
				}
				if c.Query("unsoft") != "" {
					params.UnSoft = true
				}
			} else {
				if err := c.ShouldBindBodyWith(&params, binding.JSON); err != nil {
					return err
				}
			}
			if IsBlank(params.Cond) {
				return errors.New("'cond' is required")
			}
			var multi bool
			if params.Multi || params.All {
				multi = true
			}
			_, tx = ParseCond(params.Cond, modelValue, tx)
			execDelete := func(value interface{}) error {
				bisScope := NewBizScope(c, value, tx).callCallbacks(BizDeleteKind)
				if bisScope.HasError() {
					return bisScope.DB.Error
				}
				return nil
			}
			if multi {
				result = reflect.New(reflect.SliceOf(reflectType)).Interface()
				tx = tx.Find(result)
				if tx.RowsAffected < 1 {
					return ErrAffectedDeleteToken
				}
				if params.UnSoft {
					tx = tx.Unscoped()
				}
				indirectValue := indirectValue(result)
				for index := 0; index < indirectValue.Len(); index++ {
					doc := indirectValue.Index(index).Addr().Interface()
					if err := execDelete(doc); err != nil {
						return err
					}
				}
			} else {
				result = reflect.New(reflectType).Elem().Addr().Interface()
				tx = tx.First(result)
				if tx.RowsAffected < 1 {
					return ErrAffectedDeleteToken
				}
				if params.UnSoft {
					tx = tx.Unscoped()
				}
				if err := execDelete(result); err != nil {
					return err
				}
			}
			return tx.Error
		})
		// 响应结果
		if err != nil {
			if cusErr, ok := ErrOut(err); ok {
				c.STDErr(c.L("kuu_error_"+fmt.Sprintf("%v", cusErr.Code), ErrMsgs(err)[0]), err)
			} else {
				c.STDErr(c.L("rest_delete_failed", "Delete failed"), err)
			}
		} else {
			result = Meta(reflect.New(reflectType).Interface()).OmitPassword(result)
			c.STD(result)
		}
	}
}

func restCreateHandler(reflectType reflect.Type) func(c *Context) {
	return func(c *Context) {
		var (
			docs  []interface{}
			multi bool
			err   error
		)
		// 事务执行
		err = c.WithTransaction(func(tx *gorm.DB) error {
			var body interface{}
			if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
				return err
			}
			indirectScopeValue := indirectValue(body)
			if indirectScopeValue.Kind() == reflect.Slice {
				multi = true
				for i := 0; i < indirectScopeValue.Len(); i++ {
					doc := reflect.New(reflectType).Elem().Addr().Interface()
					if err := Copy(indirectScopeValue.Index(i).Interface(), doc); err != nil {
						return err
					}
					docs = append(docs, doc)
				}
			} else {
				doc := reflect.New(reflectType).Interface()
				if err := Copy(body, doc); err != nil {
					return err
				}
				docs = append(docs, doc)
			}
			for i, doc := range docs {
				bizScope := NewBizScope(c, doc, tx).callCallbacks(BizCreateKind)
				if bizScope.HasError() {
					return bizScope.DB.Error
				}
				docs[i] = Meta(reflect.New(reflectType).Interface()).OmitPassword(doc)
			}
			return tx.Error
		})
		// 响应结果
		if err != nil {
			if cusErr, ok := ErrOut(err); ok {
				c.STDErr(c.L("kuu_error_"+fmt.Sprintf("%v", cusErr.Code), ErrMsgs(err)[0]), err)
			} else {
				c.STDErr(c.L("rest_create_failed", "Create failed"), err)
			}
		} else {
			if multi {
				c.STD(docs)
			} else {
				c.STD(docs[0])
			}
		}
	}
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
