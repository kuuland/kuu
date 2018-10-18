package rest

import (
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
		cond = body["cond"].(map[string]interface{})
		doc  = body["doc"].(map[string]interface{})
		all  = false
	)
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
	c.JSON(http.StatusOK, kuu.StdOK(data))
}
