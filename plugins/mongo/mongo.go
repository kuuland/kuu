package mongo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
	"github.com/kuuland/kuu/plugins/mongo/rest"
)

func init() {
	kuu.On("OnNew", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		if c := k.Config["mongo"]; c != nil {
			uri := c.(string)
			db.Connect(uri)
		}
	})
	kuu.On("OnModel", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		schema := args[1].(*kuu.Schema)
		rest.Mount(k, schema.Name)
	})
}

// Mount 挂载模型路由
func Mount(k *kuu.Kuu, name string) {
	rest.Mount(k, name)
}

// MetadataHandler 元数据列表路由
func MetadataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdOK(kuu.K().Schemas))
}

// SN 根据连接名获取会话
func SN(name string) *mgo.Session {
	return db.SN(name)
}

// S 获取会话
func S() *mgo.Session {
	return db.S()
}

// C 获取集合对象
func C(name string) *mgo.Collection {
	return db.C(name)
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
