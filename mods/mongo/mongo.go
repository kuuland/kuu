package mongo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

func init() {
	kuu.On("OnNew", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		if c := k.Config["mongo"]; c != nil {
			uri := c.(string)
			Connect(uri)
		}
	})
	kuu.On("OnModel", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		schema := args[1].(*kuu.Schema)
		MountAll(k, schema.Name)
	})
}

// MetadataHandler 元数据列表路由
func MetadataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdOK(kuu.K().Schemas))
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

// M 创建模型操作实例
func M(name string, args ...interface{}) *Model {
	m := &Model{
		Name: name,
	}
	if len(args) > 0 {
		m.QueryHook = args[0].(func(query *mgo.Query))
	}
	return m
}
