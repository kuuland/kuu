package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

func create(c *gin.Context) {
	// 参数处理
	var docs []interface{}
	if err := kuu.CopyBody(c, &docs); err != nil {
		handleError(err, c)
		return
	}
	// 执行查询
	C := model(name)
	defer C.Database.Session.Close()
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestCreate); ok {
		s.PreRestCreate(c, &docs)
	}
	if err := C.Insert(docs...); err != nil {
		handleError(err, c)
		return
	}
	// 触发后置钩子
	if s, ok := schema.Origin.(IPostRestCreate); ok {
		s.PostRestCreate(c, &docs)
	}
	// 构造返回
	c.JSON(http.StatusOK, kuu.StdOK(docs))
}
