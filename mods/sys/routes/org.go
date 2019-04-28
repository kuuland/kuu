package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	accounts "github.com/kuuland/kuu/mods/accounts/utils"
	"github.com/kuuland/kuu/mods/sys/models"
	"net/http"
)

// OrgLogin 组织登录
func OrgLogin() kuu.RouteInfo {
	return kuu.RouteInfo{
		Method: "POST",
		Path:   "/org/login",
		Handler: func(c *gin.Context) {
			kuu.DelGoroutineCache("LoginOrgID")
			var body map[string]string
			if err := kuu.CopyBody(c, &body); err != nil {
				kuu.Error(err)
				result := kuu.L(c, "body_parse_error")
				c.JSON(http.StatusOK, kuu.StdError(result))
				return
			}

			uid := accounts.ParseUserID(c)
			token := accounts.ParseToken(c)
			orgID := body["org_id"]

			data, err := models.ExecOrgLogin(c, orgID, uid, token)
			if err != nil {
				c.JSON(http.StatusOK, kuu.StdError(err.Error()))
				return
			}
			c.JSON(http.StatusOK, kuu.StdOK(data))
		},
	}
}

// CurrentLoginOrg 当前登录组织
func CurrentLoginOrg() kuu.RouteInfo {
	return kuu.RouteInfo{
		Method: "GET",
		Path:   "/org/current",
		Handler: func(c *gin.Context) {
			uid := accounts.ParseUserID(c)
			token := accounts.ParseToken(c)
			LoginOrg := kuu.Model("LoginOrg")
			var record models.LoginOrg
			LoginOrg.One(kuu.H{
				"Cond": kuu.H{"UID": uid, "Token": token},
				"Sort": []string{"-UpdatedAt", "-CreatedAt"},
			}, &record)
			data := kuu.H{
				"_id":  record.Org.ID,
				"Code": record.Org.Code,
				"Name": record.Org.Name,
			}
			c.JSON(http.StatusOK, kuu.StdOK(data))
		},
	}
}

// OrgList 可登录组织列表
func OrgList() kuu.RouteInfo {
	return kuu.RouteInfo{
		Method: "GET",
		Path:   "/org/list",
		Handler: func(c *gin.Context) {
			kuu.DelGoroutineCache("LoginOrgID")
			uid := accounts.ParseUserID(c)
			data := models.GetOrgList(uid)
			c.JSON(http.StatusOK, kuu.StdOK(data))
		},
	}
}
