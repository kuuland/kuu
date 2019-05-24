package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// DefaultCreateHandler
var DefaultCreateHandler = func(docs []interface{}, tx *gorm.DB, c *gin.Context) {
	sign := GetSignContext(c)
	if sign == nil || sign.OrgID == 0 {
		return
	}
	for index, doc := range docs {
		scope := tx.NewScope(doc)
		// Auto set the organization ID
		if field, exists := scope.FieldByName("OrgID"); exists {
			if err := field.Set(sign.OrgID); err != nil {
				ERROR(err)
			} else {
				docs[index] = doc
			}
		}
		// Auto set the creator ID
		if field, exists := scope.FieldByName("CreatedByID"); exists {
			if err := field.Set(sign.UID); err != nil {
				ERROR(err)
			} else {
				docs[index] = doc
			}
		}
	}
}

// DefaultDeleteHandler
var DefaultDeleteHandler = func(doc interface{}, tx *gorm.DB, c *gin.Context) {
	sign := GetSignContext(c)
	if sign == nil || sign.OrgID == 0 {
		return
	}
	scope := tx.NewScope(doc)
	// Auto set the deleter ID
	if field, exists := scope.FieldByName("DeletedByID"); exists {
		if err := field.Set(sign.UID); err != nil {
			ERROR(err)
		}
	}
}

// DefaultUpdateHandler
var DefaultUpdateHandler = func(doc interface{}, tx *gorm.DB, c *gin.Context) {
	sign := GetSignContext(c)
	if sign == nil || sign.OrgID == 0 {
		return
	}
	scope := tx.NewScope(doc)
	// Auto set the modifier ID
	if field, exists := scope.FieldByName("UpdatedByID"); exists {
		if err := field.Set(sign.UID); err != nil {
			ERROR(err)
		}
	}
}

// DefaultWhereHandler
var DefaultWhereHandler = func(db *gorm.DB, desc *PrivilegesDesc, c *gin.Context) *gorm.DB {
	if desc != nil && desc.UID != RootUID() {
		db = db.Where("(org_id IS NULL) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
	}
	return db
}
