package kuu

import "github.com/jinzhu/gorm"

// RestQueryHooks
type RestQueryHooks interface {
	RestBeforeQuery(*gorm.DB) *gorm.DB
}
