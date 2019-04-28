package models

import (
	"github.com/kuuland/kuu/mods/mongo"
)

// Role 系统角色
type Role struct {
	ID                  string                `json:"_id" displayName:"系统角色"`
	Code                string                `name:"角色编码"`
	Name                string                `name:"角色名称"`
	OperationPrivileges []OperationPrivileges `name:"操作权限"`
	DataPrivileges      []DataPrivileges      `name:"数据权限"`
	IsBuiltIn           bool                  `name:"是否系统内置"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}

// OperationPrivileges 操作权限定义
type OperationPrivileges struct {
	Permission string `name:"权限编码"`
	Desc       string `name:"权限描述"`
}

// DataPrivileges 数据权限定义
type DataPrivileges struct {
	OrgID            string `name:"组织ID"`
	OrgName          string `name:"组织名称"`
	AllReadableRange string `name:"全局可读范围" dict:"sys_data_range"`
	AllWritableRange string `name:"全局可写范围" dict:"sys_data_range"`
	AuthObjects      []AuthObject
}

// AuthObject 授权实体
type AuthObject struct {
	Name             string `name:"实体名称"`
	DisplayName      string `name:"实体显示名"`
	ObjReadableRange string `name:"实体可读范围" dict:"sys_data_range"`
	ObjWritableRange string `name:"实体可写范围" dict:"sys_data_range"`
}

// AfterSave 新增/修改后钩子
func (u *Role) AfterSave(scope *mongo.Scope) (err error) {
	UpdateAllAuthRules()
	return
}

// AfterRemove 删除后钩子
func (u *Role) AfterRemove(scope *mongo.Scope) (err error) {
	UpdateAllAuthRules()
	return
}
