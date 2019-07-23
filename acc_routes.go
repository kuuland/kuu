package kuu

import (
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"regexp"
	"time"
)

type GenTokenDesc struct {
	UID      uint
	Payload  jwt.MapClaims
	Exp      int64 `binding:"required"`
	SubDocID uint
	Desc     string `binding:"required"`
	IsAPIKey bool
}

// GenToken
func GenToken(desc GenTokenDesc) (secretData *SignSecret, err error) {
	// 设置JWT令牌信息
	iat := time.Now().Unix()
	desc.Payload["Iat"] = iat      // 签发时间
	desc.Payload["Exp"] = desc.Exp // 过期时间
	// 生成新密钥
	secretData = &SignSecret{
		UID:      desc.UID,
		Secret:   uuid.NewV4().String(),
		Iat:      iat,
		Exp:      desc.Exp,
		Method:   "LOGIN",
		SubDocID: desc.SubDocID,
		Desc:     desc.Desc,
		IsAPIKey: NewNullBool(desc.IsAPIKey),
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
	// 缓存secret至redis
	//key := RedisKeyBuilder(RedisSecretKey, secretData.Token)
	//value := Stringify(&secretData)
	//if err := RedisClient.SetNX(key, value, time.Unix(desc.Exp, 0).Sub(time.Now())).Err(); err != nil {
	//	ERROR("令牌缓存到Redis失败：%s", err.Error())
	//}
	// 保存登入历史
	saveHistory(secretData)
	return
}

// LoginRoute
var LoginRoute = RouteInfo{
	Method: "POST",
	Path:   "/login",
	HandlerFunc: func(c *Context) {
		// 调用登录处理器获取登录数据
		if loginHandler == nil {
			PANIC("login handler not configured")
		}
		payload, uid := loginHandler(c)
		if uid == 0 || len(payload) == 0 {
			return
		}
		// 调用令牌签发
		secretData, err := GenToken(GenTokenDesc{
			UID:     uid,
			Payload: payload,
			Exp:     time.Now().Add(time.Second * time.Duration(ExpiresSeconds)).Unix(),
		})
		if err != nil {
			c.STDErr(L("acc_token_failed", "Token signing failed"), err)
			return
		}
		// 设置到上下文中
		c.Set(SignContextKey, &SignContext{
			Token:   secretData.Token,
			UID:     secretData.UID,
			Payload: payload,
			Secret:  secretData,
		})
		// 设置Cookie
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
		c.STD(payload)
	},
}

// LogoutRoute
var LogoutRoute = RouteInfo{
	Method: "POST",
	Path:   "/logout",
	HandlerFunc: func(c *Context) {
		if c.SignInfo != nil && c.SignInfo.IsValid() {
			var (
				secretData SignSecret
				db         = DB()
			)
			db.Where(&SignSecret{UID: c.SignInfo.UID, Token: c.SignInfo.Token}).First(&secretData)
			if !db.NewRecord(&secretData) {
				if err := db.Model(&secretData).Updates(&SignSecret{Method: "LOGOUT"}).Error; err != nil {
					c.STDErr(L("acc_logout_failed", "Logout failed"), err)
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
	Method: "POST",
	Path:   "/valid",
	HandlerFunc: func(c *Context) {
		if c.SignInfo != nil && c.SignInfo.IsValid() {
			c.SignInfo.Payload[TokenKey] = c.SignInfo.Token
			c.STD(c.SignInfo.Payload)
		} else {
			c.STDErrHold(L("acc_token_expired", "Token has expired")).Code(555).Render()
		}
	},
}

// APIKeyRoute
var APIKeyRoute = RouteInfo{
	Method: "POST",
	Path:   "/apikeys",
	HandlerFunc: func(c *Context) {
		var body GenTokenDesc
		failedMessage := L("apikeys_failed", "Create API & Keys failed")
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
