package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// Remove 定义了删除路由接口
func Remove(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := kuu.GetSchema(name)
	handler := func(c *gin.Context) {
		scope := &Scope{
			Context: c,
			Cache:   make(kuu.H),
		}
		// 参数处理
		var body kuu.H
		if err := kuu.CopyBody(c, &body); err != nil {
			handleError(err, c)
			return
		}
		var (
			cond = kuu.H{}
			doc  = kuu.H{}
			all  = false
		)
		kuu.JSONConvert(body["cond"], &cond)
		if body["all"] != nil {
			all = body["all"].(bool)
		}
		doc = setUpdatedBy(c, doc)
		// 执行查询
		m := kuu.Model(name)
		scope.Model = m
		scope.RemoveCond = &cond
		scope.RemoveDoc = &doc
		scope.RemoveAll = all
		var (
			err  error
			data interface{}
		)
		scope.CallMethod(BeforeRemoveRouteEnum, schema)
		if all == true {
			data, err = m.RemoveAllWithData(cond, doc)
		} else {
			err = m.RemoveWithData(cond, doc)
			data = body
		}
		if err != nil {
			handleError(err, c)
			return
		}
		// 构造返回
		res := kuu.StdOK(data)
		scope.ResponseData = &res
		scope.CallMethod(AfterRemoveRouteEnum, schema)
		c.JSON(http.StatusOK, res)
	}
	return handler
}
