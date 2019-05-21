package kuu

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Model struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	OrgID       uint
	CreatedByID uint
	UpdatedByID uint
	DeletedByID uint
	Org         *Org  `gorm:"foreignkey:OrgID"`
	CreatedBy   *User `gorm:"foreignkey:CreatedByID"`
	UpdatedBy   *User `gorm:"foreignkey:UpdatedByID"`
	DeletedBy   *User `gorm:"foreignkey:DeletedByID"`
	Remark      string
}

// Metadata
type Metadata struct {
	Model     `rest:"*"`
	Name      string
	FullName  string
	Fields    []MetadataField
	IsBuiltIn bool
}

// QueryPreload
func (m *Metadata) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Fields")
}

// MetadataField
type MetadataField struct {
	Model
	Name       string
	Type       string
	MetadataID uint
}

// Routes
type Route struct {
	Model  `rest:"*"`
	Method string
	Path   string
}

// User
type User struct {
	Model       `rest:"*"`
	Username    string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	Name        string
	Avatar      string
	Sex         int
	Mobile      string
	Email       string
	Language    string
	Disable     bool
	RoleAssigns []RoleAssign
	IsBuiltIn   bool
}

// BeforeSave
func (u *User) BeforeSave() {
	if len(u.Password) == 32 {
		u.Password = GenerateFromPassword(u.Password)
	}
	return
}

// QueryPreload
func (u *User) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("RoleAssigns")
}

// Org
type Org struct {
	Model        `rest:"*"`
	Code         string `gorm:"unique;not null"`
	Name         string `gorm:"unique;not null"`
	Pid          uint
	Sort         int
	FullPathPid  string
	FullPathName string
}

type RoleAssign struct {
	Model
	UserID     uint
	RoleID     uint
	ExpiryUnix int64
}

// Role
type Role struct {
	Model               `rest:"*"`
	Code                string   `gorm:"unique;not null"`
	Name                string    `gorm:"not null"`
	OperationPrivileges []OperationPrivileges
	DataPrivileges      []DataPrivileges
	IsBuiltIn           bool
}

// AfterSave
func (r *Role) AfterSave() {
	UpdateAuthRules(nil)
}

// AfterSave
func (r *Role) AfterDelete(tx *gorm.DB) {
	UpdateAuthRules(tx)
}

// QueryPreload
func (r *Role) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("OperationPrivileges").Preload("DataPrivileges")
}

// OperationPrivileges
type OperationPrivileges struct {
	Model
	RoleID     uint
	Permission string
	Desc       string
}

// DataPrivileges
type DataPrivileges struct {
	Model
	RoleID           uint
	TargetOrg        *Org `gorm:"foreignkey:TargetOrgID"`
	TargetOrgID      uint
	AllReadableRange string
	AllWritableRange string
	AuthObjects      []AuthObject
}

// QueryPreload
func (d *DataPrivileges) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("TargetOrg").Preload("AuthObjects")
}

// AuthObject
type AuthObject struct {
	Model
	Name             string
	DisplayName      string
	ObjReadableRange string
	ObjWritableRange string
}

// Menu
type Menu struct {
	Model         `rest:"*"`
	Code          string `gorm:"unique;not null"`
	Name          string `gorm:"not null"`
	URI           string
	Icon          string
	Pid           uint
	Group         string
	Disable       bool
	IsLink        bool
	IsVirtual     bool
	Sort          int
	IsBuiltIn     bool
	IsDefaultOpen bool
	Closeable     bool
	Type          string
}

// AuthRule
type AuthRule struct {
	Model          `rest:"*"`
	UID            uint
	Username       string
	Name           string
	TargetOrgID    uint
	TargetOrg      Org `gorm:"foreignkey:TargetOrgID"`
	ObjectName     string
	ReadableScope  string
	WritableScope  string
	ReadableOrgIDs string
	WritableOrgIDs string
	HitAssign      uint
	Permissions    string
}

// Dict
type Dict struct {
	Model     `rest:"*"`
	Code      string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	Values    []DictValue
	IsBuiltIn bool
}

// QueryPreload
func (d *Dict) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Values")
}

// DictValue
type DictValue struct {
	Model
	DictID uint
	Label  string
	Value  string
	Sort   int
}

// File
type File struct {
	Model  `rest:"*"`
	UID    string `json:"uid"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
	Path   string `json:"path"`
}

// SignOrg
type SignOrg struct {
	Model
	Token string
	UID   uint
}

// IsValid
func (o *SignOrg) IsValid() bool {
	if o != nil && o.Org.ID > 0 {
		return true
	}
	return false
}

// Param
type Param struct {
	Model     `rest:"*"`
	Code      string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	Value     string
	IsBuiltIn bool
}
