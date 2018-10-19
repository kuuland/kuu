package kuu

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func ExampleNew() {
	k := New(H{
		"name": "kuu",
	})
	k.GET("/", func(c *gin.Context) {
		c.String(200, "hello")
	})
	fmt.Printf("Hello %s.\n", k.Name)
	k.Run(":8080")
	// Output:
	// Hello kuu.
}
