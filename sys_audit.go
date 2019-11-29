package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"regexp"
)

var (
	// AuditCreateCallback
	AuditCreateCallback = auditCreateCallback
	// AuditUpdateCallback
	AuditUpdateCallback = auditUpdateCallback
	// AuditDeleteCallback
	AuditDeleteCallback = auditDeleteCallback
)

func registerAuditCallbacks(callback *gorm.Callback) {
	if callback.Create().Get("kuu:audit_create") == nil {
		callback.Create().After("gorm:commit_or_rollback_transaction").Register("kuu:audit_create", AuditCreateCallback)
	}
	if callback.Update().Get("kuu:audit_update") == nil {
		callback.Update().After("gorm:commit_or_rollback_transaction").Register("kuu:audit_update", AuditUpdateCallback)
	}
	if callback.Delete().Get("kuu:audit_delete") == nil {
		callback.Update().After("gorm:commit_or_rollback_transaction").Register("kuu:audit_delete", AuditDeleteCallback)
	}
}

// NewAuditLog
func NewAuditLog(scope *gorm.Scope, auditType string) {
	if !IsNil(scope.Value) {
		if _, ok := scope.Value.(LogIgnorer); ok {
			return
		}
	}

	logTableName := scope.DB().NewScope(&Log{}).TableName()
	reg := regexp.MustCompile(fmt.Sprintf(`/INSERT\s*INTO\s*%s/i`, logTableName))
	if reg.MatchString(scope.SQL) {
		return
	}

	info := NewLog(LogTypeAudit)
	info.AuditType = auditType
	info.AuditTag = "system"
	info.AuditSQL = scope.SQL
	info.AuditSQLVars = Stringify(scope.SQLVars, false)
	if meta := Meta(scope.Value); meta != nil {
		info.AuditModel = meta.Name
	}
	info.Save2Cache()
}

func auditCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		NewAuditLog(scope, AuditTypeCreate)
	}
}

func auditUpdateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		NewAuditLog(scope, AuditTypeUpdate)
	}
}

func auditDeleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var op string
		_, hasDeletedAtField := scope.FieldByName("DeletedAt")
		if !scope.Search.Unscoped && hasDeletedAtField {
			op = AuditTypeUpdate
		} else {
			op = AuditTypeRemove
		}
		NewAuditLog(scope, op)
	}
}
