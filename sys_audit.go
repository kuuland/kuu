package kuu

import (
	"github.com/jinzhu/gorm"
	"net/http"
	"net/url"
	"time"
)

var (
	// AuditCreateCallback
	AuditCreateCallback = auditCreateCallback
	// AuditUpdateCallback
	AuditUpdateCallback = auditUpdateCallback
	// AuditDeleteCallback
	AuditDeleteCallback = auditDeleteCallback
)

type AuditInfo struct {
	Time           string      `json:",omitempty"`
	Model          string      `json:",omitempty"`
	RequestMethod  string      `json:",omitempty"`
	RequestPath    string      `json:",omitempty"`
	RequestHeaders http.Header `json:",omitempty"`
	RequestQuery   url.Values  `json:",omitempty"`
	UID            uint        `json:",omitempty"`
	SubDocID       uint        `json:",omitempty"`
	Token          string      `json:",omitempty"`
	Op             string      `json:",omitempty"`
	SQL            interface{} `json:",omitempty"`
	SQLVars        interface{} `json:",omitempty"`
}

// Output
func (info *AuditInfo) Output() {
	content := Stringify(info, C().DefaultGetBool("audit:format", true))
	INFO("KUU AUDIT: %s", content)
}

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

// NewAuditInfo
func NewAuditInfo(scope *gorm.Scope, op string) *AuditInfo {
	info := AuditInfo{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		SQL:     scope.SQL,
		SQLVars: scope.SQLVars,
		Op:      op,
	}
	if c := GetRoutineRequestContext(); c != nil {
		info.RequestHeaders = c.Request.Header
		info.RequestQuery = c.Request.URL.Query()
		info.RequestMethod = c.Request.Method
		info.RequestPath = c.Request.URL.Path
	}
	if desc := GetRoutinePrivilegesDesc(); desc != nil && desc.SignInfo != nil {
		info.UID = desc.UID
		info.SubDocID = desc.SignInfo.SubDocID
		info.Token = desc.SignInfo.Token
	}
	return &info
}

func auditCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		NewAuditInfo(scope, "CREATE").Output()
	}
}

func auditUpdateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		NewAuditInfo(scope, "UPDATE").Output()
	}
}

func auditDeleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var op string
		_, hasDeletedAtField := scope.FieldByName("DeletedAt")
		if !scope.Search.Unscoped && hasDeletedAtField {
			op = "UPDATE"
		} else {
			op = "DELETE"
		}
		NewAuditInfo(scope, op).Output()
	}
}
