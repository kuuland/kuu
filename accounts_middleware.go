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
	if whiteListCheck(c) == true {
		c.Next()
	} else {
		// 从请求参数中解码令牌
		sign, err := DecodedContext(c)
		if err != nil {
			ERROR(err)
			std := STDRender{
				Message: err.Error(),
				Code:    555,
			}
			c.AbortWithStatusJSON(http.StatusOK, &std)
			return
		}
		if sign.IsValid() {
			c.Next()
		} else {
			ERROR(err)
			std := STDRender{
				Message: err.Error(),
				Code:    555,
			}
			c.AbortWithStatusJSON(http.StatusOK, &std)
		}
	}
}
