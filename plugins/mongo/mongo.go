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

// All 插件声明
func All() *kuu.Plugin {
	return &kuu.Plugin{
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/meta",
				Handler: MetadataHandler,
			},
		},
	}
}

// Model 创建模型操作实例
func Model(name string, args ...interface{}) *M {
	m := &M{
		Name: name,
	}
	if len(args) > 0 {
		m.QueryHook = args[0].(func(query *mgo.Query))
	}
	return m
}
