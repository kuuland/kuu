package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// ID 定义了ID查询路由接口
func ID(k *kuu.Kuu, name string) func(*gin.Context) {
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
		scope.CallMethod(BeforeIDEnum, schema)
		// 构造返回
		var data kuu.H
		err := m.ID(p, &data)
		if err != nil {
			kuu.Error(err)
			c.JSON(http.StatusOK, kuu.StdError(kuu.SafeL(defaultMessages, c, "entity_not_exist")))
			return
		}
		// 构造返回
		res := kuu.StdOK(data)
		scope.ResponseData = &res
		scope.CallMethod(AfterIDEnum, schema)
		c.JSON(http.StatusOK, res)
	}
	return handler
}
