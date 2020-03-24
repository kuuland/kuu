package kuu

import (
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
	"regexp"
	"time"
)

const (
	SignMethodLogin  = "LOGIN"
	SignMethodLogout = "LOGOUT"
)

type GenTokenDesc struct {
	UID      uint
	Payload  jwt.MapClaims
	Exp      int64 `binding:"required"`
	SubDocID uint
	Desc     string `binding:"required"`
	IsAPIKey bool
	Type     string
}

// GenToken
func GenToken(desc GenTokenDesc) (secretData *SignSecret, err error) {
	// 设置JWT令牌信息
	iat := time.Now().Unix()
	desc.Payload["Iat"] = iat      // 签发时间
	desc.Payload["Exp"] = desc.Exp // 过期时间
	if desc.Type == "" {
		desc.Type = AdminSignType
	}
	// 兼容未传递SubDocID时自动查询
	if desc.SubDocID == 0 {
		var user User
		db := DB().Model(&User{}).Where(&User{ID: desc.UID})
		db = db.Select([]string{db.Dialect().Quote("id"), db.Dialect().Quote("sub_doc_id")})
		if err := db.First(&user).Error; err != nil {
			return nil, err
		}
		desc.SubDocID = user.SubDocID
	}
	// 生成新密钥
	secretData = &SignSecret{
		UID:      desc.UID,
		Secret:   uuid.NewV4().String(),
		Iat:      iat,
		Exp:      desc.Exp,
		Method:   SignMethodLogin,
		SubDocID: desc.SubDocID,
		Desc:     desc.Desc,
		Type:     desc.Type,
		IsAPIKey: null.NewBool(desc.IsAPIKey, true),
	}
	// 签发令牌
	if signed, err := EncodedToken(desc.Payload, secretData.Secret); err != nil {
		return secretData, err
	} else {
		secretData.Token = signed
	}
	desc.Payload[TokenKey] = secretData.Token
	if err = DB().Create(secretData).Error; err != nil {
		return
	}
	// 保存登入历史
	saveHistory(secretData)
	return
}

// LoginRoute
var LoginRoute = RouteInfo{
	Name:   "默认登录接口",
	Method: "POST",
	Path:   "/login",
	HandlerFunc: func(c *Context) {
		c.Set("is_login", true)
		// 调用登录处理器获取登录数据
		if loginHandler == nil {
			PANIC("login handler not configured")
		}
		resp := loginHandler(c)
		if resp.Error != nil {
			c.ERROR(resp.Error).STDErr(resp.LanguageMessage)
			return
		}
		// 调用令牌签发
		secretData, err := GenToken(GenTokenDesc{
			UID:     resp.UID,
			Payload: resp.Payload,
			Exp:     time.Now().Add(time.Second * time.Duration(ExpiresSeconds)).Unix(),
		})
		if err != nil {
			c.STDErr(c.L("acc_token_failed", "Token signing failed"), err)
			return
		}
		// 设置到上下文中
		c.Set(SignContextKey, &SignContext{
			Token:   secretData.Token,
			UID:     secretData.UID,
			Payload: resp.Payload,
			Secret:  secretData,
		})
		// 设置Cookie
		c.SetCookie(RequestLangKey, resp.Lang, ExpiresSeconds, "/", "", false, true)
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
		// 清空验证码Cookie和缓存
		c.SetCookie(CaptchaIDKey, "", -1, "/", "", false, true)
		DelCache(getFailedTimesKey(resp.Username))
		c.STD(resp.Payload)
	},
}

// LogoutRoute
var LogoutRoute = RouteInfo{
	Name:   "默认登出接口",
	Method: "POST",
	Path:   "/logout",
	HandlerFunc: func(c *Context) {
		c.Set("is_logout", true)
		if c.SignInfo != nil && c.SignInfo.IsValid() {
			var (
				secretData SignSecret
				db         = DB()
			)
			db.Where(&SignSecret{UID: c.SignInfo.UID, Token: c.SignInfo.Token}).First(&secretData)
			if !db.NewRecord(&secretData) {
				if err := db.Model(&secretData).Updates(&SignSecret{Method: SignMethodLogout}).Error; err != nil {
					c.STDErr(c.L("acc_logout_failed", "Logout failed"), err)
					return
				}
				// 保存登出历史
				saveHistory(&secretData)
				// 设置Cookie过期
				c.SetCookie(TokenKey, secretData.Token, -1, "/", "", false, true)
				c.SetCookie(RequestLangKey, "", -1, "/", "", false, true)
			}
		}
		c.STD("退出成功")
	},
}

// ValidRoute
var ValidRoute = RouteInfo{
	Name:   "令牌有效性验证接口",
	Method: "POST",
	Path:   "/valid",
	HandlerFunc: func(c *Context) {
		if c.SignInfo != nil && c.SignInfo.IsValid() {
			// 查询用户
			var user User
			if err := c.IgnoreAuth().DB().Select("lang, act_org_id").First(&user, "id = ?", c.SignInfo.UID).Error; err != nil {
				c.STDErr(c.L("user_query_failed", "Query user failed"), err)
				return
			}
			// 处理Lang参数
			if user.Lang == "" {
				user.Lang = ParseLang(c.Context)
			}
			c.SetCookie(RequestLangKey, user.Lang, ExpiresSeconds, "/", "", false, true)
			c.SignInfo.Payload["Lang"] = user.Lang
			c.SignInfo.Payload["ActOrgID"] = c.PrisDesc.ActOrgID
			c.SignInfo.Payload["ActOrgCode"] = c.PrisDesc.ActOrgCode
			c.SignInfo.Payload["ActOrgName"] = c.PrisDesc.ActOrgName
			c.SignInfo.Payload[TokenKey] = c.SignInfo.Token
			if c.PrisDesc != nil {
				c.SignInfo.Payload["Permissions"] = c.PrisDesc.Permissions
				c.SignInfo.Payload["RolesCode"] = c.PrisDesc.RolesCode
			}
			c.STD(c.SignInfo.Payload)
		} else {
			c.STDErrHold(c.L("acc_token_expired", "Token has expired")).Code(555).Render()
		}
	},
}

// APIKeyRoute
var APIKeyRoute = RouteInfo{
	Name:   "令牌生成接口",
	Method: "POST",
	Path:   "/apikeys",
	HandlerFunc: func(c *Context) {
		var body GenTokenDesc
		failedMessage := c.L("apikeys_failed", "Create API & Keys failed")
		if err := c.ShouldBindJSON(&body); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		body.Payload = c.SignInfo.Payload
		body.UID = c.SignInfo.UID
		body.SubDocID = c.SignInfo.SubDocID
		body.IsAPIKey = true
		secretData, err := GenToken(body)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		c.STD(secretData.Token)
	},
}

// WhitelistRoute
var WhitelistRoute = RouteInfo{
	Name:   "查询白名单列表",
	Method: "GET",
	Path:   "/whitelist",
	HandlerFunc: func(c *Context) {
		var list []string
		for _, item := range Whitelist {
			if v, ok := item.(string); ok {
				list = append(list, v)
			} else if v, ok := item.(*regexp.Regexp); ok {
				list = append(list, v.String())
			}
		}
		c.STD(list)
	},
}
