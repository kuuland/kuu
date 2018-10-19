package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

func id(c *gin.Context) {
	// 参数处理
	p := ParseParams(c)
	id := c.Param("id")
	// 执行查询
	C := model(name)
	defer C.Database.Session.Close()
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestID); ok {
		id = s.PreRestID(c, id, p)
	}
	query := C.FindId(bson.ObjectIdHex(id))
	if p.Project != nil {
		query.Select(p.Project)
	}
	// 构造返回
	var data kuu.H
	err := query.One(&data)
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
