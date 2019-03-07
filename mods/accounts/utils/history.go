package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/models"
)

// CreateSignHistory 新增登录历史
func CreateSignHistory(c *gin.Context, secret *models.UserSecret, logout bool) error {
	if c == nil || secret == nil {
		kuu.Warn("Empty params!")
		return nil
	}
	method := "login"
	if logout {
		method = "logout"
	}
	if secret == nil {
		if v, e := c.Get(ContextSecretKey); e {
			secret = v.(*models.UserSecret)
		}
	}
	hisData := &models.SignHistory{
		Token:  secret.Token,
		Method: method,
	}
	reqData := kuu.H{
		"Headers": c.Request.Header,
		"Query":   c.Request.URL.Query(),
	}
	body := kuu.H{}
	kuu.CopyBody(c, &body)
	reqData["Body"] = body
	hisData.ReqData = reqData
	hisData.LoginData = secret
	SignHistory := kuu.Model("SignHistory")
	if _, err := SignHistory.Create(hisData); err != nil {
		kuu.Error(err)
		return errors.New(kuu.L(c, "logout_failed", "Logout failed, please contact the administrator!"))
	}
	return nil
}
