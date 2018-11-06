package system

import "github.com/kuuland/kuu"

// All 模块导出
func All() *kuu.Mod {
	return &kuu.Mod{
		Models: []interface{}{
			&User{},
			&UserGroup{},
			&Role{},
			&RolePermissionAssign{},
			&UserGroupRoleAssign{},
			&UserRoleAssign{},
			&Permission{},
			&Resource{},
			&Menu{},
			&Button{},
			&Page{},
			&Param{},
		},
	}
}
