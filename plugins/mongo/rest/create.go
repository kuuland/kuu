package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

func create(c *gin.Context) {
	// 参数处理
	var docs []interface{}
	if err := kuu.CopyBody(c, &docs); err != nil {
		handleError(err, c)
		return
	}
	// 执行查询
	Model := db.Model{
		Name:      name,
		QueryHook: nil,
	}
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestCreate); ok {
		s.PreRestCreate(c, &docs)
	}
	if err := Model.Create(docs); err != nil {
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
