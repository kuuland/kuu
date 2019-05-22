package kuu

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"time"
)

// LoginRoute
var LoginRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/login",
	HandlerFunc: func(c *gin.Context) {
		// 调用登录处理器获取登录数据
		if loginHandler == nil {
			PANIC("login handler not configured")
		}
		payload, uid, err := loginHandler(c)
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
			UID:    uid,
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
		key := RedisKeyBuilder(RedisSecretKey, secretData.Token)
		value := Stringify(&secretData)
		if !RedisClient.SetNX(key, value, expiration).Val() {
			ERROR("Token cache failed")
		}
		// 保存登入历史
		saveHistory(c, &secretData)
		// 设置到上下文中
		c.Set(SignContextKey, &SignContext{
			Token:   secretData.Token,
			UID:     secretData.UID,
			Payload: payload,
			Secret:  &secretData,
		})
		// 设置Cookie
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
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
					STDErr(c, L(c, "退出登录失败"))
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
		STD(c, L(c, "登录成功"))
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
			sign.Payload[TokenKey] = sign.Token
			STD(c, sign.Payload)
		} else {
			STDErr(c, LFull(c, "token_expired", "Token has expired: '{{token}}'", gin.H{"token": sign.Token}), 555)
		}
	},
}
