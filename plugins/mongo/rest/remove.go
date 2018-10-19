package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

func remove(c *gin.Context) {
	// 参数处理
	var body kuu.H
	if err := kuu.CopyBody(c, &body); err != nil {
		handleError(err, c)
		return
	}
	var (
		cond = kuu.H{}
		all  = false
	)
	kuu.CloneDeep(body["cond"], &cond)
	if body["all"] != nil {
		all = body["all"].(bool)
	}
	if cond["_id"] != nil {
		cond["_id"] = bson.ObjectIdHex(cond["_id"].(string))
	}
	// 执行查询
	C := model(name)
	defer C.Database.Session.Close()
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestRemove); ok {
		s.PreRestRemove(c, &cond, all)
	}
	var (
		err  error
		data interface{}
	)
	if all == true {
		data, err = C.RemoveAll(cond)
	} else {
		err = C.Remove(cond)
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
