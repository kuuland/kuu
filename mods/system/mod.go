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
		Langs: map[string]kuu.LangMessages{
			"en": kuu.LangMessages{
				"login_error":   "Login failed.",
				"login_success": "User {{name}} login succeeded!",
				"table_del_ok":  "{{count}} records has been deleted!",
			},
			"zh_CN": kuu.LangMessages{
				"login_error":   "登录失败！",
				"login_success": "账号 {{name}} 登录成功！",
				"table_del_ok":  "已成功删除 {{count}} 条数据！",
			},
			"zh_TW": kuu.LangMessages{
				"login_error":   "登陸失敗！",
				"login_success": "賬號 {{name}} 登陸成功！",
				"table_del_ok":  "已成功刪除 {{count}} 條數據！",
			},
		},
	}
}
