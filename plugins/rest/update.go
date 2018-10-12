package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Update 修改接口
func Update(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "修改接口",
		})
	}
}
