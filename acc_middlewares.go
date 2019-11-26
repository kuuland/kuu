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
		var (
			loginMessage   = L("acc_please_login", "Please login").C(c)
			expiredMessage = L("acc_session_expired", "Login session has expired").C(c)
		)
		if err != nil {
			STDErrHold(c, loginMessage, err).Code(555).Abort()
			return
		}
		if sign.IsValid() {
			c.Next()
		} else {
			STDErrHold(c, expiredMessage, err).Code(555).Abort()
			return
		}
	}
}
