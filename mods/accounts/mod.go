package accounts

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/middleware"
	"github.com/kuuland/kuu/mods/accounts/models"
	"github.com/kuuland/kuu/mods/accounts/routes"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

// SetLoginHandler 设置登录处理函数
func SetLoginHandler(handler func(*gin.Context) (interface{}, error)) {
	utils.LoginHandler = handler
}

// SetTokenKey 设置登录令牌键
func SetTokenKey(key string) {
	utils.TokenKey = key
}

// SetUserIDKey 设置用户ID键
func SetUserIDKey(key string) {
	utils.UserIDKey = key
}

// SetWhiteList 设置白名单
func SetWhiteList(list []string, replace bool) []string {
	if list != nil && len(list) > 0 {
		z := make([]string, len(utils.WhiteList)+len(list))
		l := []([]string){}
		if replace {
			l = append(l, list)
		} else {
			l = append(l, utils.WhiteList)
			l = append(l, list)
		}
		exists := map[string]bool{}
		offset := 0
		for _, arr := range l {
			for i, item := range arr {
				if exists[item] {
					continue
				}
				exists[item] = true
				z[i+offset] = item
			}
			offset += len(arr)
		}
		utils.WhiteList = z
	}
	return utils.WhiteList
}

// All 模块声明
func All(loginHandler func(*gin.Context) (interface{}, error)) *kuu.Mod {
	if loginHandler != nil {
		SetLoginHandler(loginHandler)
	}
	return &kuu.Mod{
		Models: []interface{}{
			&models.UserSecret{},
			&models.SignHistory{},
		},
		Middleware: kuu.Middleware{
			middleware.Auth,
		},
		Routes: kuu.Routes{
			routes.LoginHandler,
			routes.LogoutHandler,
			routes.ValidHandler,
		},
		Langs: map[string]kuu.LangMessages{
			"en": kuu.LangMessages{
				"login_error": "Login failed, please contact the administrator or try again later!",
				"auth_error":  "Your session has expired, please log in again!",
				"logout":      "Logout successful!",
			},
			"zh_CN": kuu.LangMessages{
				"login_error": "登录失败，请稍后重试或联系管理员！",
				"auth_error":  "您的会话已过期，请重新登录！",
				"logout":      "登录成功！",
			},
			"zh_TW": kuu.LangMessages{
				"login_error": "登錄失敗，請稍後重試或聯繫系統管理員！",
				"auth_error":  "您的會話已過期，請重新登錄！",
				"logout":      "登錄成功！",
			},
		},
	}
}
