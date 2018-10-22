package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

func id(c *gin.Context) {
	// 参数处理
	p := ParseParams(c)
	id := p.ID
	// 执行查询
	Model := db.Model{
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
	err := Model.ID(p, &data)
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
