package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

func update(c *gin.Context) {
	// 参数处理
	var body kuu.H
	if err := kuu.CopyBody(c, &body); err != nil {
		handleError(err, c)
		return
	}
	c.Set("body", &body)
	var (
		cond = kuu.H{}
		doc  = kuu.H{}
		all  = false
	)
	if b, e := json.Marshal(body["cond"]); e == nil {
		json.Unmarshal(b, &cond)
	}
	if b, e := json.Marshal(body["doc"]); e == nil {
		json.Unmarshal(b, &doc)
	}
	if body["all"] != nil {
		all = body["all"].(bool)
	}
	if cond["_id"] != nil {
		cond["_id"] = bson.ObjectIdHex(cond["_id"].(string))
	}
	if doc["$set"] == nil {
		doc = kuu.H{
			"$set": doc,
		}
	}
	// 执行查询
	Model := db.Model{
		Name:      name,
		QueryHook: nil,
	}
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestUpdate); ok {
		s.PreRestUpdate(c, &cond, &doc, all)
	}
	var (
		err  error
		data interface{}
	)
	if all == true {
		data, err = Model.UpdateAll(cond, doc)
	} else {
		err = Model.Update(cond, doc)
		data = body
	}
	if err != nil {
		handleError(err, c)
		return
	}
	// 触发后置钩子
	if s, ok := schema.Origin.(IPostRestUpdate); ok {
		s.PostRestUpdate(c, &data)
	}
	// 构造返回
	c.JSON(http.StatusOK, kuu.StdOK(data))
}
