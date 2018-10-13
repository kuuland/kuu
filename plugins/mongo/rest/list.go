package rest

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

const (
	// ALL 全量模式
	ALL = "ALL"
	// PAGE 分页模式
	PAGE = "PAGE"
)

// List 列表查询接口
func List(name string) func(*gin.Context) {
	return func(c *gin.Context) {
		// 参数处理
		p := parseParams(c)
		// 执行查询
		C := kuu.D("mongo:C", name).(*mgo.Collection)
		defer C.Database.Session.Close()
		query := C.Find(p.cond)
		totalrecords, countErr := query.Count()
		if countErr != nil {
			handleError(countErr, c)
			return
		}
		if p.project != nil {
			query.Select(p.project)
		}
		if p._range == PAGE {
			query.Skip((p.page - 1) * p.size).Limit(p.size)
		}
		var list []kuu.H
		findErr := query.All(&list)
		if findErr != nil {
			handleError(findErr, c)
			return
		}
		query.Sort(p.sort...)
		if list == nil {
			list = make([]kuu.H, 0)
		}
		// 构造返回
		data := kuu.H{
			"list":         list,
			"totalrecords": totalrecords,
		}
		if p._range == PAGE {
			totalpages := int(math.Ceil(float64(totalrecords) / float64(p.size)))
			data["totalpages"] = totalpages
			data["page"] = p.page
			data["size"] = p.size
		}
		if p.sort != nil && len(p.sort) > 0 {
			data["sort"] = p.sort
		}
		if p.project != nil {
			data["project"] = p.project
		}
		if p.cond != nil {
			data["cond"] = p.cond
		}
		if p._range != "" {
			data["range"] = p._range
		}
		c.JSON(http.StatusOK, data)
	}
}
