package routes

import (
	"net/http"
	"time"

	"github.com/kuuland/kuu/mods/accounts/models"
	"github.com/satori/go.uuid"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

// LoginHandler 登入路由
var LoginHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/login",
	Handler: func(c *gin.Context) {
		// 调用登录处理器获取登录数据
		payload, err := utils.LoginHandler(c)
		claims := jwt.MapClaims{}
		secret := uuid.NewV4().String()
		kuu.JSONConvert(payload, &claims)
		if err != nil {
			c.JSON(http.StatusOK, kuu.StdError(err.Error()))
			return
		}
		// 设置JWT令牌信息
		iat := time.Now().Unix()
		exp := time.Now().Add(time.Second * time.Duration(utils.ExpiresSeconds)).Unix()
		claims["iat"] = iat // JWT令牌签发时间戳
		claims["exp"] = exp // JWT令牌过期时间戳
		// 生成新密钥
		secretData := &models.UserSecret{
			UserID: claims[utils.UserIDKey].(string),
			Secret: secret,
			Iat:    iat,
			Exp:    exp,
			Method: "login",
		}
		// 签发令牌
		secretData.Token = utils.Encoded(claims, secret)
		claims[utils.TokenKey] = secretData.Token
		UserSecret := kuu.Model("UserSecret")
		if _, err := UserSecret.Create(secretData); err != nil {
			kuu.Error(err)
			c.JSON(http.StatusOK, kuu.StdError("Login failed, please contact the administrator"))
			return
		}
		// 保存登入历史
		utils.CreateSignHistory(c, secretData, false)
		// 设置Cookie
		c.SetCookie(utils.TokenKey, secretData.Token, utils.ExpiresSeconds, "/", "", false, true)
		c.SetCookie(utils.UserIDKey, secretData.UserID, utils.ExpiresSeconds, "/", "", false, true)
		// 设置上下文缓存，便于其他中间件在登录后进行逻辑处理
		tokenCacheKey, tokenCacheVal := "LoginToken", secretData.Token
		uidCacheKey, uidCacheVal := "LoginUID", secretData.UserID
		c.Set(utils.ContextSecretKey, secretData)
		c.Set(utils.ContextClaimsKey, claims)
		c.Set(tokenCacheKey, tokenCacheVal)
		c.Set(uidCacheKey, uidCacheVal)
		kuu.SetGoroutineCache(tokenCacheKey, tokenCacheKey)
		kuu.SetGoroutineCache(uidCacheKey, uidCacheVal)
		c.JSON(http.StatusOK, kuu.StdOK(claims))
	},
}
