package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

// ID ID查询接口
func ID(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		// 参数处理
		p := parseParams(c)
		// 执行查询
		C := kuu.D("mongo:C", name).(*mgo.Collection)
		defer C.Database.Session.Close()
		query := C.FindId(bson.ObjectIdHex(c.Param("id")))
		if p.project != nil {
			query.Select(p.project)
		}
		// 构造返回
		var data kuu.H
		err := query.One(&data)
		if err != nil {
			handleError(err, c)
			return
		}
		c.JSON(http.StatusOK, kuu.StdDataOK(data))
	}
}
