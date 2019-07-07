package kuu

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthMiddleware
func AuthMiddleware(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	if InWhitelist(c) == true {
		c.Next()
	} else {
		// 从请求参数中解码令牌
		sign, err := DecodedContext(c)
		if err != nil {
			STDErrHold(c, "请登录后访问", err).Code(555).Abort()
			return
		}
		if sign.IsValid() {
			c.Next()
		} else {
			STDErrHold(c, "登录会话已过期", err).Code(555).Abort()
			return
		}
	}
}
