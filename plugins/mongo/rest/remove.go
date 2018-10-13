package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

// Remove 删除接口
func Remove(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		// 参数处理
		var body kuu.H
		if err := c.ShouldBindJSON(&body); err != nil {
			handleError(err, c)
			return
		}
		var (
			cond = body["cond"].(map[string]interface{})
			all  = false
		)
		if body["all"] != "" {
			all = body["all"].(bool)
		}
		if cond["_id"] != nil {
			cond["_id"] = bson.ObjectIdHex(cond["_id"].(string))
		}
		// 执行查询
		C := kuu.D("mongo:C", name).(*mgo.Collection)
		defer C.Database.Session.Close()
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
		// 构造返回
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}
