package kuu

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApp(t *testing.T) {
	k := Default()
	k.GET("/", func(c *gin.Context) {
		c.String(200, "hello")
	})
	k.Run(":8080")
}
