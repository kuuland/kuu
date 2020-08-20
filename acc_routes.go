package kuu

import (
	"github.com/jinzhu/gorm"
	"regexp"
	"time"
)

const (
	SignMethodLogin  = "LOGIN"
	SignMethodLogout = "LOGOUT"
)

// LoginRoute
var LoginRoute = RouteInfo{
	Name:   "默认登录接口",
	Method: "POST",
	Path:   "/login",
	IntlMessages: map[string]string{
		"acc_login_failed": "Login failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		// 调用登录处理器获取登录数据
		if loginHandler == nil {
			PANIC("login handler not configured")
		}
		resp := loginHandler(c)
		if resp.Error != nil {
			if resp.LocaleMessageID != "" {
				return c.STDErr(resp.Error, resp.LocaleMessageID, resp.LocaleMessageDefaultText, resp.LocaleMessageContextValues)
			}
			return c.STDErr(resp.Error, "acc_login_failed")
		}
		// 调用令牌签发
		secretData, err := GenToken(GenTokenDesc{
			UID:     resp.UID,
			Payload: resp.Payload,
			Exp:     time.Now().Add(time.Second * time.Duration(ExpiresSeconds)).Unix(),
			Type:    AdminSignType,
		})
		if err != nil {
			return c.STDErr(err, "acc_login_failed")
		}
		// 设置到上下文中
		c.Set("__kuu_sign_context__", &SignContext{
			Token:   secretData.Token,
			UID:     secretData.UID,
			Payload: resp.Payload,
			Secret:  secretData,
		})
		// 设置Cookie
		c.SetCookie(LangKey, resp.Lang, ExpiresSeconds, "/", "", false, true)
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
		// 清空验证码Cookie和缓存
		c.SetCookie(CaptchaIDKey, "", -1, "/", "", false, true)
		DelCache(getFailedTimesKey(resp.Username))
		return c.STD(resp.Payload)
	},
}

// LogoutRoute
var LogoutRoute = RouteInfo{
	Name:   "默认登出接口",
	Method: "POST",
	Path:   "/logout",
	IntlMessages: map[string]string{
		"acc_logout_failed": "Logout failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		if err := Logout(c, c.DB()); err != nil {
			return c.STDErr(err, "acc_logout_failed")
		}
		return c.STDOK()
	},
}

func Logout(c *Context, tx *gorm.DB) error {
	if c.SignInfo != nil && c.SignInfo.IsValid() {
		var secret SignSecret
		if err := tx.Where(&SignSecret{UID: c.SignInfo.UID, Token: c.SignInfo.Token}).First(&secret).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if secret.ID != 0 {
			if err := tx.Model(&secret).Updates(&SignSecret{Method: SignMethodLogout}).Error; err != nil {
				return err
			}
			// 保存登出历史
			saveHistory(&secret)
			// 设置Cookie过期
			c.SetCookie(TokenKey, secret.Token, -1, "/", "", false, true)
			c.SetCookie(LangKey, "", -1, "/", "", false, true)
		}
	}
	return nil
}

// ValidRoute
var ValidRoute = RouteInfo{
	Name:   "令牌有效性验证接口",
	Method: "POST",
	Path:   "/valid",
	IntlMessages: map[string]string{
		"acc_invalid_token": "Invalid token.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		// 查询用户
		var user User
		if err := c.IgnoreAuth().DB().Select("lang, act_org_id").First(&user, "id = ?", c.SignInfo.UID).Error; err != nil {
			return c.STDErr(err, "acc_invalid_token")
		}
		// 处理Lang参数
		if user.Lang == "" {
			user.Lang = c.Lang()
		}
		c.SetCookie(LangKey, user.Lang, ExpiresSeconds, "/", "", false, true)
		c.SignInfo.Payload["Lang"] = user.Lang
		c.SignInfo.Payload["ActOrgID"] = c.PrisDesc.ActOrgID
		c.SignInfo.Payload["ActOrgCode"] = c.PrisDesc.ActOrgCode
		c.SignInfo.Payload["ActOrgName"] = c.PrisDesc.ActOrgName
		c.SignInfo.Payload[TokenKey] = c.SignInfo.Token
		if c.PrisDesc != nil {
			c.SignInfo.Payload["Permissions"] = c.PrisDesc.Permissions
			c.SignInfo.Payload["RolesCode"] = c.PrisDesc.RolesCode
		}
		return c.STD(c.SignInfo.Payload)
	},
}

// APIKeyRoute
var APIKeyRoute = RouteInfo{
	Name:   "令牌生成接口",
	Method: "POST",
	Path:   "/apikeys",
	IntlMessages: map[string]string{
		"apikeys_failed": "Create API Keys failed.",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var body GenTokenDesc
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "apikeys_failed")
		}
		body.Payload = c.SignInfo.Payload
		body.UID = c.SignInfo.UID
		body.IsAPIKey = true
		body.Type = AdminSignType
		secretData, err := GenToken(body)
		if err != nil {
			return c.STDErr(err, "apikeys_failed")
		}
		return c.STD(secretData.Token)
	},
}

// WhitelistRoute
var WhitelistRoute = RouteInfo{
	Name:   "查询白名单列表",
	Method: "GET",
	Path:   "/whitelist",
	HandlerFunc: func(c *Context) *STDReply {
		var list []string
		for _, item := range Whitelist {
			if v, ok := item.(string); ok {
				list = append(list, v)
			} else if v, ok := item.(*regexp.Regexp); ok {
				list = append(list, v.String())
			}
		}
		return c.STD(list)
	},
}
