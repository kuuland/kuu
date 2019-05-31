package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func registerCallbacks() {
	if gorm.DefaultCallback.Query().Get("kuu:query") == nil {
		gorm.DefaultCallback.Query().Before("gorm:query").Register("kuu:query", QueryCallback)
	}
	if gorm.DefaultCallback.Query().Get("kuu:before_query") == nil {
		gorm.DefaultCallback.Query().Before("gorm:query").Register("kuu:before_query", BeforeQueryCallback)
	}
	if gorm.DefaultCallback.Update().Get("kuu:update") == nil {
		gorm.DefaultCallback.Update().Before("gorm:update").Register("kuu:update", UpdateCallback)
	}
	if gorm.DefaultCallback.Delete().Get("kuu:delete") == nil {
		gorm.DefaultCallback.Delete().Before("gorm:delete").Register("kuu:delete", DeleteCallback)
	}
	if gorm.DefaultCallback.Create().Get("kuu:create") == nil {
		gorm.DefaultCallback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
	}
	if gorm.DefaultCallback.Create().Get("kuu:create") == nil {
		gorm.DefaultCallback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
	}
}

// BeforeQueryCallback
var BeforeQueryCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("BeforeFind")
	}
}

// CreateCallback
var CreateCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		if v, ok := GetValue(PrisDescKey); ok && v != nil {
			desc := v.(*PrivilegesDesc)
			if desc.IsValid() && !scope.HasError() {
				if orgIDField, ok := scope.FieldByName("OrgID"); ok {
					if orgIDField.IsBlank {
						orgIDField.Set(desc.SignOrgID)
					}
				}
				if createdByField, ok := scope.FieldByName("CreatedByID"); ok {
					createdByField.Set(desc.UID)
				}

				if updatedByField, ok := scope.FieldByName("UpdatedByID"); ok {
					updatedByField.Set(desc.UID)
				}
			}
		}
	}
}

// DeleteCallback
var DeleteCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		if v, ok := GetValue(PrisDescKey); ok && v != nil {
			desc := v.(*PrivilegesDesc)
			if desc.IsValid() && !scope.HasError() {
				var extraOption string
				if str, ok := scope.Get("gorm:delete_option"); ok {
					extraOption = fmt.Sprint(str)
				}
				deletedByField, hasDeletedByField := scope.FieldByName("DeletedByID")
				if !scope.Search.Unscoped && hasDeletedByField {
					scope.Raw(fmt.Sprintf(
						"UPDATE %v SET %v=%v%v%v",
						scope.QuotedTableName(),
						scope.Quote(deletedByField.DBName),
						scope.AddToVars(desc.UID),
						AddExtraSpaceIfExist(scope.CombinedConditionSql()),
						AddExtraSpaceIfExist(extraOption),
					)).Exec()
				}
			}
		}
	}
}

// AddExtraSpaceIfExist
func AddExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// UpdateCallback
var UpdateCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		if v, ok := GetValue(PrisDescKey); ok && v != nil {
			desc := v.(*PrivilegesDesc)
			if desc.IsValid() {
				scope.SetColumn("UpdatedByID", desc.UID)
			}
		}
	}
}

// QueryCallback
var QueryCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		rawDesc, _ := GetValue(PrisDescKey)
		rawValues, _ := GetValue(ValuesKey)

		if !IsBlank(rawDesc) {
			if !IsBlank(rawValues) {
				values := make(Values)
				values = *(rawValues.(*Values))
				if _, ok := values[IgnoreAuthKey]; ok {
					return
				}
			}
			desc := rawDesc.(*PrivilegesDesc)
			if desc.NotRootUser() {
				_, hasOrgIDField := scope.FieldByName("OrgID")
				_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
				if hasOrgIDField && hasCreatedByIDField {
					scope.Search.Where("(org_id IS NULL) OR (org_id = 0) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
				}
			}
		}
	}
}
