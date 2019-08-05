package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

// Model
type Model struct {
	ID          uint               `gorm:"primary_key"`
	CreatedAt   time.Time          `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time          `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time         `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	OrgID       uint               `name:"所属组织ID（默认字段）"`
	CreatedByID uint               `name:"创建人ID（默认字段）"`
	UpdatedByID uint               `name:"修改人ID（默认字段）"`
	DeletedByID uint               `name:"删除人ID（默认字段）"`
	Remark      string             `name:"备注"`
	Org         *Org               `gorm:"foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User              `gorm:"foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User              `gorm:"foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User              `gorm:"foreignkey:id;association_foreignkey:deleted_by_id"`
	ExAttrs     map[string]*ExAttr `gorm:"-"`
}

// GetAttrBool
func (model *Model) HasAttr(name string) (has bool) {
	_, has = model.ExAttrs[name]
	return
}

// GetAttrBool
func (model *Model) GetAttrBool(name string) (v bool) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetBool()
	}
	return
}

// GetAttrFloat64
func (model *Model) GetAttrFloat64(name string) (v float64) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetFloat64()
	}
	return
}

// GetAttrString
func (model *Model) GetAttrString(name string) (v string) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetString()
	}
	return
}

// GetAttrInt64
func (model *Model) GetAttrInt64(name string) (v int64) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetInt64()
	}
	return
}

// GetAttrInt
func (model *Model) GetAttrInt(name string) (v int) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetInt()
	}
	return
}

// GetAttrUint
func (model *Model) GetAttrUint(name string) (v uint) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetUint()
	}
	return
}

// GetAttrUint
func (model *Model) GetAttrTime(name string) (v time.Time) {
	if model.HasAttr(name) {
		v = model.ExAttrs[name].GetTime()
	}
	return
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

// ExAttr
type ExAttr struct {
	gorm.Model
	Meta   string
	RowID  uint
	Name   string
	Kind   string
	Bool   null.Bool
	Float  null.Float
	String null.String
	Int    null.Int
	Time   null.Time
}

// GetBool
func (ex *ExAttr) GetBool() bool {
	return ex.Bool.Bool
}

// GetFloat64
func (ex *ExAttr) GetFloat64() float64 {
	return ex.Float.Float64
}

// GetString
func (ex *ExAttr) GetString() string {
	return ex.String.String
}

// GetInt64
func (ex *ExAttr) GetInt64() int64 {
	return ex.Int.Int64
}

// GetInt
func (ex *ExAttr) GetInt() int {
	return int(ex.GetInt())
}

// GetUint
func (ex *ExAttr) GetUint() uint {
	return uint(ex.GetInt())
}

// GetTime
func (ex *ExAttr) GetTime() time.Time {
	return ex.Time.Time
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
	Model `rest:"*" displayName:"用户" kuu:"password"`
	ExtendField
	Username    string       `name:"账号" gorm:"not null"`
	Password    string       `name:"密码" gorm:"not null" json:",omitempty"`
	Name        string       `name:"姓名"`
	Avatar      string       `name:"头像"`
	Sex         int          `name:"性别"`
	Mobile      string       `name:"手机号"`
	Email       string       `name:"邮箱地址"`
	Disable     null.Bool    `name:"是否禁用"`
	RoleAssigns []RoleAssign `name:"已分配角色"`
	IsBuiltIn   null.Bool    `name:"是否内置"`
	SubDocID    uint         `name:"扩展档案ID"`
	Lang        string       `name:"最近使用语言"`
	AllowLogin  bool         `name:"允许登录"`
}

// BeforeSave
func (u *User) BeforeSave(scope *gorm.Scope) (err error) {
	if len(u.Password) == 32 {
		var hashed string
		if hashed, err = GenerateFromPassword(u.Password); err == nil {
			err = scope.SetColumn("Password", hashed)
		}
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
		childrenMap = make(map[uint][]Org)
		fullPidMap  = make(map[uint]string)
		fullNameMap = make(map[uint]string)
		fall        func([]Org, uint)
	)
	for _, org := range list {
		if org.Pid != org.ID {
			childrenMap[org.Pid] = append(childrenMap[org.Pid], org)
		}
	}
	fall = func(values []Org, pid uint) {
		for _, item := range values {
			if _, exists := fullPidMap[pid]; exists {
				fullPidMap[item.ID] = fmt.Sprintf("%s,%d", fullPidMap[pid], item.ID)
				fullNameMap[item.ID] = fmt.Sprintf("%s,%s", fullNameMap[pid], item.Name)
			} else {
				fullPidMap[item.ID] = fmt.Sprintf("%d", item.ID)
				fullNameMap[item.ID] = fmt.Sprintf("%s", item.Name)
			}
			children := childrenMap[item.ID]
			if len(children) > 0 {
				fall(children, item.ID)
			}
		}
	}
	fall(childrenMap[0], 0)
	for index, item := range list {
		item.FullPid = fullPidMap[item.ID]
		item.FullName = fullNameMap[item.ID]
		list[index] = item
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
	UserID     uint `name:"用户ID" gorm:"not null"`
	RoleID     uint `name:"角色ID" gorm:"not null"`
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
	IsBuiltIn           null.Bool             `name:"是否内置"`
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
	ReadableRange string `name:"可读范围" enum:"DataScope"`
	WritableRange string `name:"可写范围" enum:"DataScope"`
}

// Menu
type Menu struct {
	ModelExOrg `rest:"*" displayName:"菜单"`
	ExtendField
	Code          string    `name:"菜单编码" gorm:"not null"`
	Name          string    `name:"菜单名称" gorm:"not null"`
	URI           string    `name:"菜单地址"`
	Icon          string    `name:"菜单图标"`
	Pid           uint      `name:"父菜单ID"`
	Group         string    `name:"菜单分组名"`
	Disable       null.Bool `name:"是否禁用"`
	IsLink        null.Bool `name:"是否外链"`
	Sort          int       `name:"排序值"`
	IsBuiltIn     null.Bool `name:"是否内置"`
	IsDefaultOpen null.Bool `name:"是否默认打开"`
	Closeable     null.Bool `name:"是否可关闭"`
	LocaleKey     string    `name:"国际化语言键"`
	IsVirtual     null.Bool
	Type          string
}

func updatePresetRolePrivileges(tx *gorm.DB, deleteBefore bool, ignoreAuth bool) {
	if ignoreAuth {
		IgnoreAuth()
	}
	if deleteBefore {
		tx.Unscoped().Where(OperationPrivileges{RoleID: RootRoleID()}).Delete(OperationPrivileges{})
	}
	var menus []Menu
	tx.Find(&menus)
	for _, menu := range menus {
		tx.Create(&OperationPrivileges{
			Model: Model{
				CreatedByID: RootUID(),
				UpdatedByID: RootUID(),
				OrgID:       RootOrgID(),
			},
			RoleID:   RootRoleID(),
			MenuCode: menu.Code,
		})
	}
	if ignoreAuth {
		IgnoreAuth(true)
	}
}

// AfterSave
func (m *Menu) BeforeSave() {
	if m.Code == "" {
		if m.URI != "" && !m.IsLink.Bool {
			code := m.URI
			if strings.HasPrefix(m.URI, "/") {
				code = m.URI[1:]
			}
			code = strings.ReplaceAll(code, "/", ":")
			m.Code = code
		} else {
			m.Code = RandCode()
		}
	}
}

// AfterSave
func (m *Menu) AfterSave(db *gorm.DB) {
	updatePresetRolePrivileges(db, true, true)
}

// AfterDelete
func (m *Menu) AfterDelete(db *gorm.DB) {
	updatePresetRolePrivileges(db, true, true)
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
	Token string `gorm:"size:4096"`
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
	Code      string    `name:"参数编码" gorm:"not null"`
	Name      string    `name:"参数名称" gorm:"not null"`
	Value     string    `name:"参数值"`
	IsBuiltIn null.Bool `name:"是否预置"`
}
