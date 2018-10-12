package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Remove 删除接口
func Remove(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "删除接口",
		})
	}
}
