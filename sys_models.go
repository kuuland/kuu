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

//TableName 设置表名
func (Metadata) TableName() string {
	return "sys_Metadata"
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

//TableName 设置表名
func (MetadataField) TableName() string {
	return "sys_MetadataField"
}

// Routes
type Route struct {
	Model  `rest:"*"`
	Method string
	Path   string
}

//TableName 设置表名
func (Route) TableName() string {
	return "sys_Route"
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

//TableName 设置表名
func (User) TableName() string {
	return "sys_User"
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
	Model    `rest:"*"`
	Code     string `gorm:"unique;not null"`
	Name     string `gorm:"unique;not null"`
	Pid      uint
	Sort     int
	FullPid  string
	FullName string
	Class    string
}

// AfterSave
func (o *Org) AfterSave() {
	delPrisCache()
}

// AfterDelete
func (o *Org) AfterDelete() {
	delPrisCache()
}

//TableName 设置表名
func (Org) TableName() string {
	return "sys_Org"
}

type RoleAssign struct {
	Model      `rest:"*"`
	UserID     uint
	RoleID     uint
	Role       *Role
	ExpireUnix int64
}

// AfterSave
func (u *User) AfterSave() {
	delPrisCache()
}

// AfterDelete
func (u *User) AfterDelete() {
	delPrisCache()
}

//TableName 设置表名
func (RoleAssign) TableName() string {
	return "sys_RoleAssign"
}

// Role
type Role struct {
	Model               `rest:"*"`
	Code                string `gorm:"unique;not null"`
	Name                string `gorm:"not null"`
	OperationPrivileges []OperationPrivileges
	DataPrivileges      []DataPrivileges
	IsBuiltIn           bool
}

//TableName 设置表名
func (Role) TableName() string {
	return "sys_Role"
}

// AfterSave
func (r *Role) AfterSave() {
	delPrisCache()
}

// AfterDelete
func (r *Role) AfterDelete() {
	delPrisCache()
}

// QueryPreload
func (r *Role) QueryPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("OperationPrivileges").Preload("DataPrivileges")
}

// OperationPrivileges
type OperationPrivileges struct {
	Model    `rest:"*"`
	RoleID   uint
	MenuCode string
}

//TableName 设置表名
func (OperationPrivileges) TableName() string {
	return "sys_OperationPrivileges"
}

// DataPrivileges
type DataPrivileges struct {
	Model         `rest:"*"`
	RoleID        uint
	TargetOrgID   uint
	ReadableRange string
	WritableRange string
}

//TableName 设置表名
func (DataPrivileges) TableName() string {
	return "sys_DataPrivileges"
}

// Menu
type Menu struct {
	Model         `rest:"*"`
	Code          string
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

//TableName 设置表名
func (Menu) TableName() string {
	return "sys_Menu"
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

//TableName 设置表名
func (AuthRule) TableName() string {
	return "sys_AuthRule"
}

// Dict
type Dict struct {
	Model     `rest:"*"`
	Code      string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	Values    []DictValue
	IsBuiltIn bool
}

//TableName 设置表名
func (Dict) TableName() string {
	return "sys_Dict"
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

//TableName 设置表名
func (DictValue) TableName() string {
	return "sys_DictValue"
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

//TableName 设置表名
func (File) TableName() string {
	return "sys_File"
}

// SignOrg
type SignOrg struct {
	Model
	Token string
	UID   uint
}

//TableName 设置表名
func (SignOrg) TableName() string {
	return "sys_SignOrg"
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

//TableName 设置表名
func (Param) TableName() string {
	return "sys_Param"
}
