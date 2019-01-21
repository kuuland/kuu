package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// LogoutHandler 登出路由
var LogoutHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/logout",
	Handler: func(c *gin.Context) {
		c.JSON(http.StatusOK, kuu.StdOK("ok"))
	},
}
