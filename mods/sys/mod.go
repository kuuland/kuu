package sys

import (
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/sys/middleware"
	"github.com/kuuland/kuu/mods/sys/models"
	"github.com/kuuland/kuu/mods/sys/routes"
)

func init() {
	kuu.On("OnModel", func(args ...interface{}) {
		schema := args[0].(*kuu.Schema)
		fields := *schema.Fields
		var orgField kuu.SchemaField
		for _, field := range fields {
			if field.Code == "Org" {
				orgField = field
				break
			}
		}
		if orgField.Code == "" {
			orgField = kuu.SchemaField{
				Code: "Org",
				Name: "所属组织",
				Type: "string",
				Tags: map[string]string{
					"join": "Org<Code,Name>",
				},
				JoinName: "Org",
				JoinSelect: map[string]int{
					"Code": 1,
					"Name": 1,
				},
			}
			fields = append(fields, orgField)
		}
		schema.Fields = &fields
	})
	kuu.On("BeforeRun", func(data ...interface{}) {
		sysInit()
		hooksInit()
		models.UpdateAllAuthRules()
		cacheRootID()
	})
}

func cacheRootID() {
	User := kuu.Model("User")
	var rootUser models.User
	User.One(kuu.H{
		"Cond": kuu.H{
			"Username": "root",
		},
		"Project": map[string]int{"_id": 1},
	}, &rootUser)
	if rootUser.ID != "" {
		kuu.Data["RootUID"] = rootUser.ID
	}
}

// All 模块导出
func All() *kuu.Mod {
	return &kuu.Mod{
		Models: []interface{}{
			&models.Audit{},
			&models.AuthRule{},
			&models.Dict{},
			&models.File{},
			&models.I18n{},
			&models.Menu{},
			&models.Org{},
			&models.LoginOrg{},
			&models.Param{},
			&models.Role{},
			&models.User{},
			&models.Message{},
		},
		Middleware: kuu.Middleware{
			middleware.Org,
		},
		Routes: kuu.Routes{
			routes.Init(),
			routes.Upload(),
			routes.OrgLogin(),
			routes.CurrentLoginOrg(),
			routes.OrgList(),
			routes.UserRoles(),
		},
		Langs: map[string]kuu.LangMessages{
			"en": {
				"login_user_not_exist":  "The login user profile does not exist. Please contact the administrator.",
				"login_user_not_assign": "The logged in user does not contain any valid roles.",
				"body_parse_error":      "Request parameter parsing failed.",
				"role_query_error":      "Role query failed.",
				"org_id_not_found":      "Organization ID not found in request.",
				"org_not_exist":         "Organization does not exist.",
				"org_login_error":       "Organization login failed.",
			},
			"zh_CN": {
				"login_user_not_exist":  "登录用户档案不存在，请联系管理员！",
				"login_user_not_assign": "登录用户未包含任何有效角色！",
				"body_parse_error":      "请求参数解析失败！",
				"role_query_error":      "角色查询失败，请联系管理员！",
				"org_id_not_found":      "请求中未找到组织ID！",
				"org_not_exist":         "未找到匹配的组织档案！",
				"org_login_error":       "组织登录失败！",
			},
			"zh_TW": {
				"login_user_not_exist":  "登錄用戶不存在，請聯繫管理員！",
				"login_user_not_assign": "登錄用戶未包含任何有效角色！",
				"body_parse_error":      "請求參數解析失敗！",
				"role_query_error":      "角色查詢失敗，請聯繫管理員！",
				"org_id_not_found":      "請求中未找到組織ID！",
				"org_not_exist":         "未找到匹配的組織檔案！",
				"org_login_error":       "組織登錄失敗！",
			},
		},
	}
}
