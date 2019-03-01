package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// Update 定义了修改路由接口
func Update(k *kuu.Kuu, name string) func(*gin.Context) {
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
		kuu.JSONConvert(body["doc"], &doc)
		if body["all"] != nil {
			all = body["all"].(bool)
		}
		doc = setUpdatedBy(c, doc)
		// 执行查询
		m := kuu.Model(name)
		scope.Model = m
		scope.UpdateCond = &cond
		scope.UpdateDoc = &doc
		scope.UpdateAll = all
		var (
			err  error
			data interface{}
		)
		scope.CallMethod(BeforeUpdateRouteEnum, schema)
		if all == true {
			data, err = m.UpdateAll(cond, doc)
		} else {
			err = m.Update(cond, doc)
			data = body
		}
		if err != nil {
			handleError(err, c)
			return
		}
		// 构造返回
		res := kuu.StdOK(data)
		scope.ResponseData = &res
		scope.CallMethod(AfterUpdateRouteEnum, schema)
		c.JSON(http.StatusOK, res)
	}
	return handler
}
