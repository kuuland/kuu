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
	if err := C.Insert(docs...); err != nil {
		handleError(err, c)
		return
	}
	// 构造返回
	c.JSON(http.StatusOK, kuu.StdOK(docs))
}
