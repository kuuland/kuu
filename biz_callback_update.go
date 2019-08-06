package kuu

import (
	"reflect"
	"time"
)

func init() {
	DefaultCallback.Update().Register("kuu:biz_before_update", bizBeforeUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_update", bizUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_after_update", bizAfterUpdateCallback)
}

func bizBeforeUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeUpdate")
	}
}

func bizUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		createOrUpdateItem := func(item interface{}) {
			tx := scope.DB
			if tx.NewRecord(item) {
				if err := tx.Create(item).Error; err != nil {
					_ = scope.Err(err)
					return
				}
			} else {
				itemScope := tx.NewScope(item)
				if field, ok := itemScope.FieldByName("DeletedAt"); ok && !field.IsBlank {
					if err := itemScope.SetColumn(field.DBName, nil); err != nil {
						ERROR("清空 DeletedAt 字段失败：%s", err.Error())
					}
					if err := tx.Delete(item).Error; err != nil {
						_ = scope.Err(err)
						return
					}
				} else {
					if err := tx.Model(item).Update(item).Error; err != nil {
						_ = scope.Err(err)
						return
					}
				}
			}
		}
		dbScope := scope.DB.NewScope(scope.Value)
		for key, _ := range scope.UpdateParams.Doc {
			if field, ok := dbScope.FieldByName(key); ok {
				if field.Relationship != nil {
					switch field.Relationship.Kind {
					case "has_many", "many_to_many":
						for i := 0; i < field.Field.Len(); i++ {
							createOrUpdateItem(field.Field.Index(i).Addr().Interface())
						}
					case "has_one", "belongs_to":
						createOrUpdateItem(field.Field.Addr().Interface())
					}
					delete(scope.UpdateParams.Doc, key)
					scope.UpdateParams.Doc["UpdatedAt"] = time.Now()
				}
			}
		}
		if len(scope.UpdateParams.Doc) > 0 {
			scope.Value = reflect.New(scope.ReflectType).Interface()
			if err := Copy(scope.UpdateParams.Doc, scope.Value); err != nil {
				_ = scope.Err(err)
				return
			}
			scope.DB = scope.DB.Model(scope.UpdateCond).
				Set("gorm:association_autoupdate", false).
				Updates(scope.Value)
			if err := scope.DB.Error; err != nil {
				_ = scope.Err(err)
			} else if scope.DB.RowsAffected < 1 {
				_ = scope.Err(ErrAffectedSaveToken)
				return
			}
		}
	}
}

func bizAfterUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterUpdate")
	}
}
