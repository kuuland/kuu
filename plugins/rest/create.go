package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Create 新增接口
func Create(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "新增接口",
		})
	}
}
