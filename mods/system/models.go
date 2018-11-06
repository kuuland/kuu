package system

import "time"

const (
	// RoleTypeOrg 组织类角色
	RoleTypeOrg = iota
	// RoleTypePosition 岗位类角色
	RoleTypePosition
	// RoleTypeOther 其他类角色
	RoleTypeOther
)
const (
	// ResourceTypeMenu 菜单类资源
	ResourceTypeMenu = iota
	// ResourceTypePage 页面类资源
	ResourceTypePage
	// ResourceTypeButton 按钮类资源
	ResourceTypeButton
)

const (
	// ResourceEffectFunc 功能类资源
	ResourceEffectFunc = iota
	// ResourceEffectData 数据类资源
	ResourceEffectData
)

// Model 定义了数据模型的一些基本字段
type Model struct {
	CreatedBy interface{} `name:"创建人"`
	CreatedAt time.Time   `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人"`
	UpdatedAt time.Time   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

// User 用户数据模型
type User struct {
	Username string    `name:"账号" required:"true"`
	Password string    `name:"密码" required:"true"`
	Name     string    `name:"姓名"`
	Birth    time.Time `name:"出生日期"`
	Avatar   string    `name:"头像"`
	Sex      int       `name:"性别"`
	Disable  bool      `name:"是否禁用"`
	Mobile   string    `name:"手机号"`
	Email    string    `name:"邮箱账号"`
}

// UserGroup 用户组数据模型
type UserGroup struct {
	Name  string `name:"组名称" required:"true"`
	Users []User `name:"关联用户"`
}

// Role 角色数据模型
type Role struct {
	Name        string       `name:"角色名称" required:"true"`
	Permissions []Permission `name:"关联权限"`
}

// Resource 资源数据模型
type Resource struct {
	Name   string `name:"资源名称" required:"true"`
	Type   int    `name:"资源类型"`
	Effect int    `name:"资源作用"`
}

// Menu 菜单数据模型
type Menu struct {
	Name    string `name:"菜单名称" required:"true"`
	URI     string `name:"菜单地址" required:"true"`
	Icon    string `name:"菜单图标"`
	Parent  string `name:"父级菜单"`
	Disable bool   `name:"是否禁用"`
	IsLink  bool   `name:"是否外链"`
	Order   int    `name:"菜单排序"`
}

// Page 页面数据模型
type Page struct {
	Name string `name:"页面名称" required:"true"`
	URI  string `name:"页面地址" required:"true"`
	Desc string `name:"详细描述"`
}

// Button 按钮数据模型
type Button struct {
	Name string `name:"按钮名称" required:"true"`
	Desc string `name:"详细描述"`
}

// Permission 权限数据模型
type Permission struct {
	Name     string     `name:"权限名称" required:"true"`
	Type     string     `name:"权限类型" required:"true"`
	Resource []Resource `name:"关联资源表"`
}

// UserRoleAssign 用户角色分配记录
type UserRoleAssign struct {
	User User `name:"分配用户" required:"true"`
	Role Role `name:"分配角色" required:"true"`
}

// UserGroupRoleAssign 用户组角色分配记录
type UserGroupRoleAssign struct {
	Group UserGroup `name:"分配用户组" required:"true"`
	Role  Role      `name:"分配角色" required:"true"`
}

// RolePermissionAssign 角色权限分配记录
type RolePermissionAssign struct {
	Role       Role       `name:"分配角色" required:"true"`
	Permission Permission `name:"分配权限" required:"true"`
}

// Param 系统参数数据模型
type Param struct {
	Code  string      `name:"参数编码" required:"true"`
	Name  string      `name:"参数名称" required:"true"`
	Value interface{} `name:"参数值" required:"true"`
}
