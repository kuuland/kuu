package mongo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

var k *kuu.Kuu

func init() {
	kuu.On("OnNew", func(args ...interface{}) {
		k = args[0].(*kuu.Kuu)
		if c := k.Config["mongo"]; c != nil {
			uri := c.(string)
			Connect(uri)
		}
	})
}

// MetadataHandler 元数据列表路由
func MetadataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdOK(kuu.Schemas))
}

// All 模块声明
func All() *kuu.Mod {
	return &kuu.Mod{
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/meta",
				Handler: MetadataHandler,
			},
		},
	}
}
