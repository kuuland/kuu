package kuu

import (
	"fmt"
	"strings"
	"time"
)

// Model
type Model struct {
	ID          uint       `gorm:"primary_key"`
	CreatedAt   time.Time  `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time  `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	OrgID       uint       `name:"所属组织ID（默认字段）"`
	CreatedByID uint       `name:"创建人ID（默认字段）"`
	UpdatedByID uint       `name:"修改人ID（默认字段）"`
	DeletedByID uint       `name:"删除人ID（默认字段）"`
	Org         *Org       `gorm:"foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User      `gorm:"foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User      `gorm:"foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User      `gorm:"foreignkey:id;association_foreignkey:deleted_by_id"`
	Remark      string
}

// ModelExOrg
type ModelExOrg struct {
	ID          uint       `gorm:"primary_key"`
	CreatedAt   time.Time  `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time  `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	CreatedByID uint       `name:"创建人ID（默认字段）"`
	UpdatedByID uint       `name:"修改人ID（默认字段）"`
	DeletedByID uint       `name:"删除人ID（默认字段）"`
	CreatedBy   *User      `gorm:"foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User      `gorm:"foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User      `gorm:"foreignkey:id;association_foreignkey:deleted_by_id"`
	Remark      string
}

// ExtendField
type ExtendField struct {
	Def1 uint   `name:"扩展字段1（默认字段）"`
	Def2 string `name:"扩展字段2（默认字段）"`
	Def3 string `name:"扩展字段3（默认字段）"`
	Def4 string `name:"扩展字段4（默认字段）"`
	Def5 string `name:"扩展字段5（默认字段）"`
}

// Routes
type Route struct {
	Model `rest:"*"`
	ExtendField
	Method string
	Path   string
}

// User
type User struct {
	Model `rest:"*" displayName:"用户"`
	ExtendField
	Username    string       `name:"账号" gorm:"not null"`
	Password    string       `name:"密码" gorm:"not null" json:",omitempty" kuu:"password"`
	Name        string       `name:"姓名"`
	Avatar      string       `name:"头像"`
	Sex         int          `name:"性别"`
	Mobile      string       `name:"手机号"`
	Email       string       `name:"邮箱地址"`
	Language    string       `name:"语言"`
	Disable     bool         `name:"是否禁用"`
	RoleAssigns []RoleAssign `name:"已分配角色"`
	IsBuiltIn   bool         `name:"是否内置"`
	SubDocID    uint         `name:"扩展档案ID"`
}

// BeforeSave
func (u *User) BeforeSave() {
	if len(u.Password) == 32 {
		u.Password = GenerateFromPassword(u.Password)
	}
	return
}

// Org
type Org struct {
	Model `rest:"*" displayName:"组织"`
	ExtendField
	Code     string `name:"组织编码" gorm:"not null"`
	Name     string `name:"组织名称" gorm:"not null"`
	Pid      uint   `name:"父组织ID"`
	Sort     int    `name:"排序值"`
	FullPid  string
	FullName string
	Class    string
}

// OrgIDMap
func OrgIDMap(list []Org) map[uint]Org {
	orgMap := make(map[uint]Org)
	for _, org := range list {
		orgMap[org.ID] = org
	}
	return orgMap
}

// FillOrgFullInfo
func FillOrgFullInfo(list []Org) []Org {
	type info struct {
		fullPid  string
		fullName string
	}
	var (
		infoMap     = make(map[uint]info)
		childrenMap = make(map[uint][]Org)
		fall        func([]Org, string, string)
	)
	for _, org := range list {
		childrenMap[org.Pid] = append(childrenMap[org.Pid], org)
	}
	fall = func(values []Org, pid, pname string) {
		for _, item := range values {
			if pid != "" {
				item.FullPid = fmt.Sprintf("%s,%d", pid, item.ID)
				item.FullName = fmt.Sprintf("%s,%s", pname, item.Name)
			} else {
				item.FullPid = fmt.Sprintf("%d", item.ID)
				item.FullName = fmt.Sprintf("%s", item.Name)
			}
			if _, has := infoMap[item.ID]; !has {
				infoMap[item.ID] = info{
					fullPid:  item.FullPid,
					fullName: item.FullName,
				}
			}
			children := childrenMap[item.ID]
			if len(children) > 0 {
				fall(children, item.FullPid, item.FullName)
			}
		}
	}
	fall(list, "", "")
	for index, item := range list {
		list[index].FullPid = infoMap[item.ID].fullPid
		list[index].FullName = infoMap[item.ID].fullName
	}
	return list
}

// AfterSave
func (o *Org) AfterSave() {
	DelPrisCache()
}

// AfterDelete
func (o *Org) AfterDelete() {
	DelPrisCache()
}

// RoleAssign
type RoleAssign struct {
	Model `rest:"*" displayName:"用户角色分配"`
	ExtendField
	UserID     uint `name:"用户ID"`
	RoleID     uint `name:"角色ID"`
	Role       *Role
	ExpireUnix int64
}

// AfterSave
func (u *User) AfterSave() {
	DelPrisCache()
}

// AfterDelete
func (u *User) AfterDelete() {
	DelPrisCache()
}

// Role
type Role struct {
	Model `rest:"*" displayName:"角色"`
	ExtendField
	Code                string                `name:"角色编码" gorm:"not null"`
	Name                string                `name:"角色名称" gorm:"not null"`
	OperationPrivileges []OperationPrivileges `name:"角色操作权限"`
	DataPrivileges      []DataPrivileges      `name:"角色数据权限"`
	IsBuiltIn           bool                  `name:"是否内置"`
}

// AfterSave
func (r *Role) AfterSave() {
	DelPrisCache()
}

// AfterDelete
func (r *Role) AfterDelete() {
	DelPrisCache()
}

// OperationPrivileges
type OperationPrivileges struct {
	Model `rest:"*" displayName:"角色操作权限"`
	ExtendField
	RoleID   uint   `name:"角色ID"`
	MenuCode string `name:"菜单编码"`
}

// DataPrivileges
type DataPrivileges struct {
	Model `rest:"*" displayName:"角色数据权限"`
	ExtendField
	RoleID        uint   `name:"角色ID"`
	TargetOrgID   uint   `name:"目标组织ID"`
	ReadableRange string `name:"可读范围"`
	WritableRange string `name:"可写范围"`
}

// Menu
type Menu struct {
	ModelExOrg `rest:"*" displayName:"菜单"`
	ExtendField
	Code          string `name:"菜单编码"`
	Name          string `name:"菜单名称" gorm:"not null"`
	URI           string `name:"菜单地址"`
	Icon          string `name:"菜单图标"`
	Pid           uint   `name:"父菜单ID"`
	Group         string `name:"菜单分组名"`
	Disable       bool   `name:"是否禁用"`
	IsLink        bool   `name:"是否外链"`
	Sort          int    `name:"排序值"`
	IsBuiltIn     bool   `name:"是否内置"`
	IsDefaultOpen bool   `name:"是否默认打开"`
	Closeable     bool   `name:"是否可关闭"`
	IsVirtual     bool
	Type          string
}

// AfterSave
func (m *Menu) BeforeSave() {
	if m.Code == "" {
		if m.URI != "" && !m.IsLink {
			code := m.URI
			if strings.HasPrefix(m.URI, "/") {
				code = m.URI[1:]
			}
			code = strings.ReplaceAll(code, "/", ":")
			m.Code = code
		}
	}
}

// Dict
type Dict struct {
	Model `rest:"*"`
	ExtendField
	Code      string `gorm:"not null"`
	Name      string `gorm:"not null"`
	Values    []DictValue
	IsBuiltIn bool
}

// DictValue
type DictValue struct {
	Model
	ExtendField
	DictID uint
	Label  string
	Value  string
	Sort   int
}

// File
type File struct {
	Model `rest:"*" displayName:"文件"`
	ExtendField
	Class  string `name:"文件分类" `
	RefID  uint   `name:"关联ID" `
	UID    string `name:"文件唯一ID" `
	Type   string `name:"文件Mine-Type" `
	Size   int64  `name:"文件大小" `
	Name   string `name:"文件名称" `
	Status string `name:"文件状态" `
	URL    string `name:"文件URL" `
	Path   string `json:"path"`
}

// SignOrg
type SignOrg struct {
	Model
	ExtendField
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
	Model `rest:"*" displayName:"参数"`
	ExtendField
	Code      string `name:"参数编码" gorm:"not null"`
	Name      string `name:"参数名称" gorm:"not null"`
	Value     string `name:"参数值"`
	IsBuiltIn bool   `name:"是否预置"`
}
