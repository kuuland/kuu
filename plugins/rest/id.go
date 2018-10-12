package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ID ID查询接口
func ID(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "ID查询接口",
		})
	}
}
