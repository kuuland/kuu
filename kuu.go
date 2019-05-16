package kuu

import (
	"github.com/gin-gonic/gin"
)

// Default
func Default() *gin.Engine {
	onInit()
	return gin.Default()
}

// New
func New() *gin.Engine {
	onInit()
	return gin.New()
}

func onInit() {
	initDataSources()
	initRedis()
}
