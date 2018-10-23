package rest

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

// Update 定义了修改路由接口
func Update(k *kuu.Kuu, name string) func(*gin.Context) {
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
		kuu.JSONConvert(body["cond"], &cond)
		kuu.JSONConvert(body["doc"], &doc)
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
		var jwtData jwt.MapClaims
		if value, exists := c.Get("JWTDecoded"); exists && value != nil {
			jwtData = value.(jwt.MapClaims)
		}
		if doc["UpdatedBy"] == nil && jwtData != nil && jwtData["_id"] != nil {
			doc["UpdatedBy"] = jwtData["_id"]
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
	return handler
}
