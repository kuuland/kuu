package kuu

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"time"
)

// LoginHandler
var LoginRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/login",
	HandlerFunc: func(c *gin.Context) {
		// 调用登录处理器获取登录数据
		data, err := LoginHandler(c)
		payload := *data
		if err != nil {
			STDErr(c, err.Error())
			return
		}
		// 设置JWT令牌信息
		expiration := time.Second * time.Duration(ExpiresSeconds)
		iat := time.Now().Unix()
		exp := time.Now().Add(expiration).Unix()
		payload["iat"] = iat // 签发时间
		payload["exp"] = exp // 过期时间
		// 生成新密钥
		secretData := SignSecret{
			UID:    payload[UIDKey].(uint),
			Secret: uuid.NewV4().String(),
			Iat:    iat,
			Exp:    exp,
			Method: "LOGIN",
		}
		// 签发令牌
		secretData.Token = EncodedToken(payload, secretData.Secret)
		payload[TokenKey] = secretData.Token
		DB().Create(&secretData)
		// 缓存secret至redis
		if err := saveToRedis(&secretData, expiration); err != nil {
			ERROR(err)
		}
		// 保存登入历史
		saveHistory(c, &secretData)
		// 设置Cookie
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
		c.SetCookie(UIDKey, string(secretData.UID), ExpiresSeconds, "/", "", false, true)
		STD(c, payload)
	},
}

// LogoutRoute
var LogoutRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/logout",
	HandlerFunc: func(c *gin.Context) {
		// 从上下文缓存中读取认证信息
		var sign *SignContext
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
		}
		if sign.IsValid() {
			var (
				secretData SignSecret
				db         = DB()
			)
			db.Where(&SignSecret{UID: sign.UID, Token: sign.Token}).First(&secretData)
			if !db.NewRecord(&secretData) {
				if errs := db.Model(&secretData).Updates(&SignSecret{Method: "LOGOUT"}).GetErrors(); len(errs) > 0 {
					ERROR(errs)
					STDErr(c, L(c, "Logout failed"))
					return
				}
				// 删除redis缓存
				if err := deleteFromRedis(&secretData); err != nil {
					ERROR(err)
				}
				// 保存登出历史
				saveHistory(c, &secretData)
				// 设置Cookie过期
				c.SetCookie(TokenKey, secretData.Token, -1, "/", "", false, true)
				c.SetCookie(UIDKey, string(secretData.UID), -1, "/", "", false, true)
			}
		}
		STD(c, L(c, "Logout success"))
	},
}

// ValidRoute
var ValidRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/valid",
	HandlerFunc: func(c *gin.Context) {
		var sign *SignContext
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
		}
		if sign.IsValid() {
			STD(c, sign.Payload)
		} else {
			STDErr(c, LFull(c, "token_expired", "Token has expired: '{{token}}'", gin.H{"token": sign.Token}), 555)
		}
	},
}
