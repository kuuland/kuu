package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// Define default callbacks
func init() {
	gorm.DefaultCallback.Query().Before("gorm:query").Register("kuu:query", QueryCallback)
	gorm.DefaultCallback.Update().Before("gorm:update").Register("kuu:update", UpdateCallback)
	gorm.DefaultCallback.Delete().After("gorm:delete").Register("kuu:delete", DeleteCallback)
	gorm.DefaultCallback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
}

// CreateCallback
var CreateCallback = func(scope *gorm.Scope) {
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

// DeleteCallback
var DeleteCallback = func(scope *gorm.Scope) {
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

// AddExtraSpaceIfExist
func AddExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// UpdateCallback
var UpdateCallback = func(scope *gorm.Scope) {
	if v, ok := GetValue(PrisDescKey); ok && v != nil {
		desc := v.(*PrivilegesDesc)
		if desc.IsValid() {
			scope.SetColumn("UpdatedByID", desc.UID)
		}
	}
}

// QueryCallback
var QueryCallback = func(scope *gorm.Scope) {
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
				scope.Search.Where("(org_id IS NULL) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
			}
		}
	}
}
