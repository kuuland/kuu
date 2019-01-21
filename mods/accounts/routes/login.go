package routes

import (
	"net/http"

	"github.com/kuuland/kuu/mods/accounts/models"
	uuid "github.com/satori/go.uuid"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

// LoginHandler 登入路由
var LoginHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/login",
	Handler: func(c *gin.Context) {
		userData, err := utils.LoginHandler(c)
		payload := jwt.MapClaims{}
		secret := uuid.NewV4().String()
		kuu.JSONConvert(userData, payload)
		if err != nil {
			c.JSON(http.StatusOK, kuu.StdError(err.Error()))
		}
		secretData := &models.UserSecret{
			UserID: payload[utils.UserIDKey].(string),
			Secret: secret,
			Token:  utils.Encoded(payload, secret),
		}
		payload[utils.TokenKey] = secretData.Token
		UserModel := kuu.Model("UserModel")
		if _, err := UserModel.Create(secretData); err != nil {
			kuu.Error(err)
			c.JSON(http.StatusOK, kuu.StdError("Login failed, please contact the administrator"))
			return
		}
		// c.SetCookie(utils.TokenKey, secretData.Token, )
		c.JSON(http.StatusOK, kuu.StdOK(payload))
	},
}
