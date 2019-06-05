package kuu

import (
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"time"
)

type GenTokenDesc struct {
	UID        uint
	Payload    jwt.MapClaims
	Expiration time.Duration
	SubDocID   uint
	Desc       string
	IsAPIKey   bool
}

// GenToken
func GenToken(desc GenTokenDesc) (secretData *SignSecret, err error) {
	// 设置JWT令牌信息
	iat := time.Now().Unix()
	exp := time.Now().Add(desc.Expiration).Unix()
	desc.Payload["Iat"] = iat // 签发时间
	desc.Payload["Exp"] = exp // 过期时间
	// 生成新密钥
	secretData = &SignSecret{
		UID:      desc.UID,
		Secret:   uuid.NewV4().String(),
		Iat:      iat,
		Exp:      exp,
		Method:   "LOGIN",
		SubDocID: desc.SubDocID,
		Desc:     desc.Desc,
		IsAPIKey: desc.IsAPIKey,
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
	key := RedisKeyBuilder(RedisSecretKey, secretData.Token)
	value := Stringify(&secretData)
	if err := RedisClient.SetNX(key, value, desc.Expiration).Err(); err != nil {
		ERROR("令牌缓存到Redis失败：%s", err.Error())
	}
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
		payload, uid, err := loginHandler(c)
		if err != nil {
			// Note: 登录处理器错误直接返回错误信息
			c.STDErr(err.Error())
			return
		}
		// 调用令牌签发
		expiration := time.Second * time.Duration(ExpiresSeconds)
		secretData, err := GenToken(GenTokenDesc{
			UID:        uid,
			Payload:    payload,
			Expiration: expiration,
		})
		if err != nil {
			c.STDErrHold("令牌签发失败").Data(err).Render()
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
				if errs := db.Model(&secretData).Updates(&SignSecret{Method: "LOGOUT"}).GetErrors(); len(errs) > 0 {
					c.STDErrHold("退出登录失败").Data(errs).Render()
					return
				}
				// 删除redis缓存
				if _, err := RedisClient.Del(RedisKeyBuilder(RedisSecretKey, secretData.Token)).Result(); err != nil {
					ERROR(err)
				}
				if _, err := RedisClient.Del(RedisKeyBuilder(RedisOrgKey, secretData.Token)).Result(); err != nil {
					ERROR(err)
				}
				// 保存登出历史
				saveHistory(&secretData)
				// 设置Cookie过期
				c.SetCookie(TokenKey, secretData.Token, -1, "/", "", false, true)
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
			c.STDErr("令牌已过期", 555)
		}
	},
}

// APIKeyRoute
var APIKeyRoute = RouteInfo{
	Method: "POST",
	Path:   "/apikey",
	HandlerFunc: func(c *Context) {
		var body = struct {
			Exp  int64  `binding:"required"`
			Desc string `binding:"required"`
		}{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.STDErrHold("解析请求体失败").Data(err).Render()
			return
		}
		secretData, err := GenToken(GenTokenDesc{
			Payload:    c.SignInfo.Payload,
			UID:        c.SignInfo.UID,
			Expiration: time.Unix(body.Exp, 0).Sub(time.Now()),
			Desc:       body.Desc,
			IsAPIKey:   true,
		})
		if err != nil {
			c.STDErrHold("令牌签发失败").Data(err).Render()
			return
		}
		c.STD(secretData.Token)
	},
}
