package mongo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

// ID 定义了ID查询路由接口
func ID(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := k.Schemas[name]
	handler := func(c *gin.Context) {
		// 参数处理
		p := ParseParams(c)
		id := p.ID
		// 执行查询
		m := Model{
			Name: name,
			QueryHook: func(query *mgo.Query) {
				// 触发前置钩子
				if s, ok := schema.Origin.(IPreRestID); ok {
					id = s.PreRestID(c, id, p)
				}
			},
		}
		// 构造返回
		var data kuu.H
		err := m.ID(p, &data)
		if err != nil {
			handleError(err, c)
			return
		}
		// 触发后置钩子
		if s, ok := schema.Origin.(IPostRestID); ok {
			s.PostRestID(c, &data)
		}
		c.JSON(http.StatusOK, kuu.StdOK(data))
	}
	return handler
}
