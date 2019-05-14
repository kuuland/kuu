package kuu

import "github.com/gin-gonic/gin"

// Default
func Default() *gin.Engine {
	initDataSources()
	return gin.Default()
}

// New
func New() *gin.Engine {
	initDataSources()
	return gin.New()
}
