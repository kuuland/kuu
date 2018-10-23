package mongo

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

// Remove 定义了删除路由接口
func Remove(k *kuu.Kuu, name string) func(*gin.Context) {
	schema := k.Schemas[name]
	handler := func(c *gin.Context) {
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
		if b, e := json.Marshal(body["cond"]); e == nil {
			json.Unmarshal(b, &cond)
		}
		if body["all"] != nil {
			all = body["all"].(bool)
		}
		if cond["_id"] != nil {
			cond["_id"] = bson.ObjectIdHex(cond["_id"].(string))
		}
		// 执行查询
		Model := M{
			Name:      name,
			QueryHook: nil,
		}
		// 触发前置钩子
		if s, ok := schema.Origin.(IPreRestRemove); ok {
			s.PreRestRemove(c, &cond, all)
		}
		var (
			err  error
			data interface{}
		)
		doc = setUpdatedBy(c, doc)
		if all == true {
			data, err = Model.RemoveAll(cond, doc)
		} else {
			err = Model.Remove(cond, doc)
			data = body
		}

		if err != nil {
			handleError(err, c)
			return
		}
		// 触发后置钩子
		if s, ok := schema.Origin.(IPostRestRemove); ok {
			s.PostRestRemove(c, &data)
		}
		// 构造返回
		c.JSON(http.StatusOK, kuu.StdOK(data))
	}
	return handler
}
