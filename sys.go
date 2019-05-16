package kuu

import "time"

type Model struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	OrgID       uint
	CreatedByID uint
	UpdatedByID uint
	DeletedByID uint
	//Org         Org  `gorm:"foreignkey:OrgID"`
	//CreatedBy   User `gorm:"foreignkey:CreatedByID"`
	//UpdatedBy   User `gorm:"foreignkey:UpdatedByID"`
	//DeletedBy   User `gorm:"foreignkey:DeletedByID"`
	Remark string
}

// User
type User struct {
	Model
	Username  string
	Password  string
	Name      string
	Avatar    string
	Sex       int
	Mobile    string
	Email     string
	Language  string
	Disable   bool
	Roles     []Role `gorm:"many2many:user_roles;"`
	IsBuiltIn bool
}

// Role
type Role struct {
	Model
	Code           string
	Name           string
	Menus          []Menu `gorm:"many2many:role_menus;"`
	DataPrivileges []DataPrivileges
	IsBuiltIn      bool
}

// Org
type Org struct {
	Model
	Code     string
	Name     string
	ParentID uint
	//Parent       Org `gorm:"foreignkey:ParentID"`
	Sort         int
	FullPathPid  string
	FullPathName string
}

// DataPrivileges
type DataPrivileges struct {
	Model
	RoleID           uint
	Role             Role `gorm:"foreignkey:RoleID"`
	AllReadableRange string
	AllWritableRange string
	AuthObjects      []AuthObject
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
	Model
	Code          string
	Name          string
	URI           string
	Icon          string
	Pid           string
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

// Sys
func Sys() *Mod {
	return &Mod{}
}
