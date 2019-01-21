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
	data := utils.DecodedContext(c)
	c.Set("UserData", data)
	if whiteListCheck(c) == true || (data != nil && data.Valid() == nil) {
		c.Next()
	} else {
		if data == nil {
			kuu.Error("Token decoding failed")
		} else {
			kuu.Error("Token decoding failed: %s", data.Valid().Error())
		}
		result := kuu.SafeL(utils.DefaultMessages, c, "auth_error")
		c.AbortWithStatusJSON(http.StatusOK, kuu.StdErrorWithCode(result, 555))
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
