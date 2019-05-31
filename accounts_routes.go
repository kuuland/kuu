package kuu

import (
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"time"
)

// GenSignSecret
func GenSignSecret(payload jwt.MapClaims, uid uint, expire time.Time) (*SignSecret, error) {
	// 设置JWT令牌信息
	iat := time.Now().Unix()
	exp := expire.Unix()
	payload["iat"] = iat // 签发时间
	payload["exp"] = exp // 过期时间
	// 生成新密钥
	secretData := &SignSecret{
		UID:    uid,
		Secret: uuid.NewV4().String(),
		Iat:    iat,
		Exp:    exp,
		Method: "LOGIN",
	}
	// 签发令牌
	if signed, err := EncodedToken(payload, secretData.Secret); err != nil {
		return secretData, err
	} else {
		secretData.Token = signed
	}
	payload[TokenKey] = secretData.Token
	return secretData, nil
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
		exp := time.Now().Add(expiration)
		secretData, err := GenSignSecret(payload, uid, exp)
		DB().Create(secretData)
		if err != nil {
			c.STDErrHold("令牌签发失败").Data(err).Render()
		}
		// 缓存secret至redis
		key := RedisKeyBuilder(RedisSecretKey, secretData.Token)
		value := Stringify(&secretData)
		if err := RedisClient.SetNX(key, value, expiration).Err(); err != nil {
			ERROR("令牌缓存到Redis失败：%s", err.Error())
		}
		// 保存登入历史
		saveHistory(c, secretData)
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
				saveHistory(c, &secretData)
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
		secretData, err := GenSignSecret(c.SignInfo.Payload, c.SignInfo.UID, time.Unix(body.Exp, 0))
		if err != nil {
			c.STDErrHold("令牌签发失败").Data(err).Render()
			return
		}
		secretData.Desc = body.Desc
		secretData.IsAPIKey = true
		DB().Create(secretData)
		c.STD(secretData.Token)
	},
}
