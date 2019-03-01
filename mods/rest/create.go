package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// Create 定义了新增路由接口
func Create(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := kuu.GetSchema(name)
	handler := func(c *gin.Context) {
		scope := &Scope{
			Context: c,
			Cache:   make(kuu.H),
		}
		// 参数处理
		var docs []kuu.H
		if err := kuu.CopyBody(c, &docs); err != nil {
			handleError(err, c)
			return
		}
		docs = setCreatedBy(c, docs)
		for index, doc := range docs {
			docs[index] = setUpdatedBy(c, doc)
		}
		// 执行查询
		m := kuu.Model(name)
		scope.Model = m
		scope.CreateData = &docs
		scope.CallMethod(BeforeCreateRouteEnum, schema)
		data, err := m.Create(docs)
		if err != nil {
			handleError(err, c)
			return
		}
		kuu.JSONConvert(&data, scope.CreateData)
		// 构造返回
		res := kuu.StdOK(data)
		scope.ResponseData = &res
		scope.CallMethod(AfterCreateRouteEnum, schema)
		c.JSON(http.StatusOK, res)
	}
	return handler
}
