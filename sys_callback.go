package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// CreateCallback
var CreateCallback = func(scope *gorm.Scope) {
	if v, ok := GetValue(PrisDescKey); ok {
		desc := v.(*PrivilegesDesc)
		if desc != nil && !scope.HasError() {
			if orgIDField, ok := scope.FieldByName("OrgID"); ok {
				if orgIDField.IsBlank {
					orgIDField.Set(desc.OrgID)
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
	if v, ok := GetValue(PrisDescKey); ok {
		desc := v.(*PrivilegesDesc)
		if desc != nil && !scope.HasError() {
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
					addExtraSpaceIfExist(scope.CombinedConditionSql()),
					addExtraSpaceIfExist(extraOption),
				)).Exec()
			}
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// UpdateCallback
var UpdateCallback = func(scope *gorm.Scope) {
	if v, ok := GetValue(PrisDescKey); ok {
		desc := v.(*PrivilegesDesc)
		if desc != nil {
			scope.SetColumn("UpdatedByID", desc.UID)
		}
	}
}

// QueryCallback
var QueryCallback = func(scope *gorm.Scope) {
	if v, ok := GetValue(PrisDescKey); ok {
		desc := v.(*PrivilegesDesc)
		if desc != nil {
			scope.Search.Where("(org_id IS NULL) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
		}
	}
}
