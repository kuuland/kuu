package routes

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/models"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

// ValidHandler 校验路由
var ValidHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/valid",
	Handler: func(c *gin.Context) {
		// 从上下文缓存中读取认证信息
		var (
			claims jwt.MapClaims
			secret *models.UserSecret
		)
		if v, e := c.Get(utils.ContextClaimsKey); e {
			claims = v.(jwt.MapClaims)
		}
		if v, e := c.Get(utils.ContextSecretKey); e {
			secret = v.(*models.UserSecret)
		}
		// 如果缓存不存在，尝试重新解析一次
		if claims == nil || secret == nil {
			claims, secret = utils.DecodedContext(c)
		}
		if claims != nil && claims.Valid() == nil {
			claims[utils.TokenKey] = secret.Token
			c.JSON(http.StatusOK, kuu.StdOK(claims))
		} else {
			c.JSON(http.StatusOK, kuu.StdErrorWithCode("Token has expired", 555))
		}
	},
}
