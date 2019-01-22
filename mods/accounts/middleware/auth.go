package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/utils"
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
			c.Set(utils.ContextSecretKey, secret)
			c.Set(utils.ContextClaimsKey, claims)
			c.Next()
		} else {
			if claims == nil {
				kuu.Error("Token decoding failed")
			} else {
				kuu.Error("Token decoding failed: %s", valid.Error())
			}
			result := kuu.L(c, "auth_error")
			c.AbortWithStatusJSON(http.StatusOK, kuu.StdErrorWithCode(result, 555))
		}
	}
}

func whiteListCheck(c *gin.Context) bool {
	ret := false
	if len(utils.WhiteList) == 0 {
		return ret
	}
	value := kuu.Join(c.Request.Method, " ", c.Request.URL.Path)
	value = strings.ToLower(value)
	for _, item := range utils.WhiteList {
		if s := strings.Split(item, " "); len(s) == 1 {
			item = kuu.Join(c.Request.Method, " ", item)
		}
		item = strings.ToLower(item)
		if value == item {
			ret = true
			break
		}
	}
	return ret
}
