package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/models"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

// LogoutHandler 登出路由
var LogoutHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/logout",
	Handler: func(c *gin.Context) {
		// 从上下文缓存中读取认证信息
		var secret *models.UserSecret
		if v, e := c.Get(utils.ContextSecretKey); e {
			secret = v.(*models.UserSecret)
		}
		// 如果缓存不存在，尝试重新解析一次
		if secret == nil {
			_, secret = utils.DecodedContext(c)
		}
		if secret == nil {
			c.JSON(http.StatusOK, kuu.StdError("Token decoding failed, please contact the administrator!"))
			return
		}
		// 更新用户密钥
		UserSecret := kuu.Model("UserSecret")
		secretData := &models.UserSecret{
			UserID: secret.UserID,
			Secret: secret.Secret,
			Token:  secret.Token,
			Iat:    secret.Iat,
			Exp:    secret.Exp,
			Method: "logout",
		}
		if _, err := UserSecret.Create(secretData); err != nil {
			kuu.Error(err)
			c.JSON(http.StatusOK, kuu.StdError("Logout failed, please contact the administrator!"))
			return
		}
		// 更新登录历史
		hisData := &models.SignHistory{
			Token:  secret.Token,
			Method: "logout",
		}
		reqData := kuu.H{
			"Headers": c.Request.Header,
			"Query":   c.Request.URL.Query(),
		}
		body := kuu.H{}
		kuu.CopyBody(c, body)
		reqData["Body"] = body
		if v, err := json.Marshal(reqData); err == nil {
			hisData.ReqData = string(v)
		}
		if v, err := json.Marshal(secret); err == nil {
			hisData.LoginData = string(v)
		}
		SignHistory := kuu.Model("SignHistory")
		if _, err := SignHistory.Create(hisData); err != nil {
			kuu.Error(err)
			c.JSON(http.StatusOK, kuu.StdError("Logout failed, please contact the administrator!"))
			return
		}
		// 设置Cookie过期
		c.SetCookie(utils.TokenKey, secret.Token, -1, "/", "", false, true)
		c.SetCookie(utils.UserIDKey, secret.UserID, -1, "/", "", false, true)
		c.JSON(http.StatusOK, kuu.StdOK("Logout successful!"))
	},
}
