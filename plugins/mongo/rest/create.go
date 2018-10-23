package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

// Create 定义了新增路由接口
func Create(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := k.Schemas[name]
	handler := func(c *gin.Context) {
		// 参数处理
		var docs []kuu.H
		if err := kuu.CopyBody(c, &docs); err != nil {
			handleError(err, c)
			return
		}
		docs = setCreatedBy(c, docs)
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
	return handler
}
