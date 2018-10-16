package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

// Create 新增接口
func Create(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		// 参数处理
		var docs []interface{}
		if err := c.ShouldBindJSON(&docs); err != nil {
			handleError(err, c)
			return
		}
		c.Set("body", &docs)
		// 执行查询
		C := kuu.D("mongo:C", name).(*mgo.Collection)
		defer C.Database.Session.Close()
		if err := C.Insert(docs...); err != nil {
			handleError(err, c)
			return
		}
		// 构造返回
		c.JSON(http.StatusOK, gin.H{
			"data": docs,
		})
	}
}
