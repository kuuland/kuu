package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

func id(c *gin.Context) {
	// 参数处理
	p := parseParams(c)
	// 执行查询
	C := model(name)
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
	c.JSON(http.StatusOK, kuu.StdOK(data))
}
