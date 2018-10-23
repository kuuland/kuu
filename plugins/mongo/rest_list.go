package mongo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

// List 定义了列表查询路由接口
func List(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := k.Schemas[name]
	handler := func(c *gin.Context) {
		// 参数处理
		p := ParseParams(c)
		// 执行查询
		Model := M{
			Name: name,
			QueryHook: func(query *mgo.Query) {
				// 触发前置钩子
				if s, ok := schema.Origin.(IPreRestList); ok {
					s.PreRestList(c, query, p)
				}
			},
		}
		var list []kuu.H
		data, err := Model.List(p, &list)
		if err != nil {
			handleError(err, c)
			return
		}
		// 触发后置钩子
		if s, ok := schema.Origin.(IPostRestList); ok {
			s.PostRestList(c, &data)
		}
		c.JSON(http.StatusOK, kuu.StdOK(data))
	}
	return handler
}
