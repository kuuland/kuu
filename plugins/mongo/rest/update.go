package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
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
	C := model(name)
	defer C.Database.Session.Close()
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestUpdate); ok {
		s.PreRestUpdate(c, &cond, &doc, all)
	}
	var (
		err  error
		data interface{}
	)
	if all == true {
		data, err = C.UpdateAll(cond, doc)
	} else {
		err = C.Update(cond, doc)
		data = body
	}

	if err != nil {
		handleError(err, c)
		return
	}
	// 构造返回
	// 触发前置钩子
	if s, ok := schema.Origin.(IPostRestUpdate); ok {
		s.PostRestUpdate(c, &data)
	}
	c.JSON(http.StatusOK, kuu.StdOK(data))
}
