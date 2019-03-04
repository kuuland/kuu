package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/utils"
	"net/http"
	"regexp"
)

// Auth 认证中间件
func Auth(c *gin.Context) {
	if whiteListCheck(c) == true {
		c.Next()
	} else {
		// 从请求参数中解码令牌
		claims, secret := utils.DecodedContext(c)
		valid := claims.Valid()
		if claims != nil && valid == nil {
			// 更新令牌和密钥到上下文缓存中
			tokenCacheKey, tokenCacheVal := "LoginToken", utils.ParseToken(c)
			uidCacheKey, uidCacheVal := "LoginUID", utils.ParseUserID(c)

			c.Set(utils.ContextSecretKey, secret)
			c.Set(utils.ContextClaimsKey, claims)
			c.Set(tokenCacheKey, tokenCacheVal)
			c.Set(uidCacheKey, uidCacheVal)
			kuu.SetGoroutineCache(tokenCacheKey, tokenCacheKey)
			kuu.SetGoroutineCache(uidCacheKey, uidCacheVal)

			c.Next()

			kuu.ClearGoroutineCache()
		} else {
			if claims == nil {
				kuu.Error("Token decoding failed")
			} else {
				kuu.Error("Token decoding failed: %s", valid.Error())
			}
			result := kuu.L(c, "auth_error", "Your session may have expired, please try signing in again")
			c.AbortWithStatusJSON(http.StatusOK, kuu.StdErrorWithCode(result, 555))
		}
	}
}

func whiteListCheck(c *gin.Context) bool {
	if len(utils.WhiteList) == 0 {
		return false
	}
	input := kuu.Join(c.Request.Method, " ", c.Request.URL.Path)
	for _, item := range utils.WhiteList {
		reg := regexp.MustCompile(item)
		if reg.MatchString(input) {
			return true
		}
	}
	return false
}
