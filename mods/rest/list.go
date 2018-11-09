package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// List 定义了列表查询路由接口
func List(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := kuu.GetSchema(name)
	handler := func(c *gin.Context) {
		scope := &Scope{
			Context: c,
		}
		// 参数处理
		p := ParseParams(c)
		// 执行查询
		m := kuu.Model(name)
		scope.Params = p
		scope.Model = m
		scope.CallMethod(BeforeListRouteEnum, schema)
		var list []kuu.H
		data, err := m.List(p, &list)
		if err != nil {
			handleError(err, c)
			return
		}
		// 构造返回
		res := kuu.StdOK(data)
		scope.ResponseData = &res
		scope.CallMethod(AfterListRouteEnum, schema)
		c.JSON(http.StatusOK, res)
	}
	return handler
}
