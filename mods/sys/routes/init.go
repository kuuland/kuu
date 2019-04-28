package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// Init 系统初始化路由
func Init() kuu.RouteInfo {
	return kuu.RouteInfo{
		Method: "GET",
		Path:   "/sys/init",
		Handler: func(c *gin.Context) {
			c.JSON(http.StatusOK, kuu.StdOK("ok"))
		},
	}
}
