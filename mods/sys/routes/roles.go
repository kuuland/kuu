package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	accounts "github.com/kuuland/kuu/mods/accounts/utils"
	"github.com/kuuland/kuu/mods/sys/models"
	"net/http"
)

// UserRoles 获取用户角色列表
func UserRoles() kuu.RouteInfo {
	return kuu.RouteInfo{
		Method: "GET",
		Path:   "/user/roles",
		Handler: func(c *gin.Context) {
			uid := c.Query("uid")
			if uid == "" {
				uid = accounts.ParseUserID(c)
			}
			roles, _ := models.GetUserRoles(uid)
			if roles == nil {
				roles = make([]models.Role, 0)
			}
			data := make([]kuu.H, 0)
			for _, item := range roles {
				data = append(data, kuu.H{
					"_id":  item.ID,
					"Code": item.Code,
					"Name": item.Name,
				})
			}
			c.JSON(http.StatusOK, kuu.StdOK(data))
		},
	}
}
