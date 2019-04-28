package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	accounts "github.com/kuuland/kuu/mods/accounts/utils"
	"github.com/kuuland/kuu/mods/sys/models"
	"net/http"
	"regexp"
)

// Org 组织信息查询中间件
func Org(c *gin.Context) {
	uid := accounts.ParseUserID(c)
	token := accounts.ParseToken(c)
	LoginOrg := kuu.Model("LoginOrg")
	record := &models.LoginOrg{}
	LoginOrg.One(kuu.H{
		"Cond": kuu.H{"UID": uid, "Token": token},
		"Sort": []string{"-UpdatedAt", "-CreatedAt"},
	}, record)
	if record.ID != "" && record.Org.ID != "" {
		orgIDCacheKey, orgIDCacheVal := "LoginOrgID", record.Org.ID
		c.Set(orgIDCacheKey, orgIDCacheVal)
		kuu.SetGoroutineCache(orgIDCacheKey, orgIDCacheVal)
	}
	c.Next()
	reg := regexp.MustCompile("/api/login")
	if reg.MatchString(c.Request.RequestURI) {
		err := autoOrgLogin(c)
		if err != nil {
			kuu.Error(err)
			c.AbortWithStatusJSON(http.StatusOK, kuu.StdError(kuu.L(c, "auto_org_login_error", "Org login failed")))
		}
	}
	kuu.ClearGoroutineCache()
}

func autoOrgLogin(c *gin.Context) error {
	token := c.GetString("LoginToken")
	uid := c.GetString("LoginUID")
	if token == "" || uid == "" {
		return nil
	}
	list := models.GetOrgList(uid)
	if list != nil && len(list) == 1 && list[0] != nil {
		item := list[0]
		if v, ok := item["_id"].(string); ok && v != "" {
			_, err := models.ExecOrgLogin(c, v, uid, token)
			if err == nil {
				c.SetCookie("Org", v, 0, "", "", false, false)
			}
			return err
		}
	}
	return nil
}
