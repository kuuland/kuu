package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// ValidHandler 校验路由
var ValidHandler = kuu.RouteInfo{
	Method: "POST",
	Path:   "/valid",
	Handler: func(c *gin.Context) {
		c.JSON(http.StatusOK, kuu.StdOK("ok"))
	},
}
