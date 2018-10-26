package kuu

import "time"

// Model 定义了数据模型的一些基本字段
type Model struct {
	CreatedBy interface{} `name:"创建人"`
	CreatedAt time.Time   `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人"`
	UpdatedAt time.Time   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
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

// Role 角色数据模型
type Role struct {
	Name        string       `name:"角色名称" required:"true"`
	Permissions []Permission `name:"包含权限" required:"true"`
}

// Permission 权限数据模型
type Permission struct {
	Name string `name:"权限名称" required:"true"`
	Type string `name:"权限类型" required:"true"`
}

// UserRoles 用户角色分配记录
type UserRoles struct {
	User  User   `name:"分配用户" required:"true"`
	Roles []Role `name:"分配角色" required:"true"`
}

// RolePermissions 角色权限分配记录
type RolePermissions struct {
	Role        Role         `name:"关联角色" required:"true"`
	Permissions []Permission `name:"关联权限" required:"true"`
}

// Menu 菜单数据模型
type Menu struct {
	Name       string     `name:"菜单名称" required:"true"`
	uri        string     `name:"菜单地址" required:"true"`
	icon       string     `name:"菜单图标"`
	parent     string     `name:"父级菜单"`
	Disable    bool       `name:"是否禁用"`
	IsLink     bool       `name:"是否外链"`
	Order      int        `name:"菜单排序"`
	Permission Permission `name:"菜单权限"`
}

// Param 系统参数数据模型
type Param struct {
	Code  string `name:"参数编码" required:"true"`
	Name  string `name:"参数名称" required:"true"`
	Value string `name:"参数值" required:"true"`
}
