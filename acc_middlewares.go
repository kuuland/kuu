package kuu

import (
	"fmt"
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
			loginMessage          = L("acc_please_login", "Please login")
			expiredMessage        = L("acc_session_expired", "Login session has expired")
			incorrectTokenMessage = L("acc_incorrect_token", "Incorrect token type")
		)
		if err != nil {
			STDErrHold(c, loginMessage, err).Code(555).Abort()
			return
		}
		if sign.IsValid() {
			if !validSignType(c, sign) {
				STDErrHold(c, incorrectTokenMessage).Code(556).Abort()
				return
			}
			c.Next()
		} else {
			STDErrHold(c, expiredMessage, err).Code(555).Abort()
			return
		}
	}
}

func validSignType(c *gin.Context, sign *SignContext) bool {
	k := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
	info := routesMap[k]

	if len(info.SignType) == 0 {
		return true
	}

	for _, t := range info.SignType {
		if t == sign.Type {
			return true
		}
	}

	return false
}
