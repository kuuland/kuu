package rest

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

func list(c *gin.Context) {
	// 参数处理
	p := ParseParams(c)
	// 执行查询
	C := model(name)
	defer C.Database.Session.Close()
	query := C.Find(p.Cond)
	totalrecords, countErr := query.Count()
	if countErr != nil {
		handleError(countErr, c)
		return
	}
	if p.Project != nil {
		query.Select(p.Project)
	}
	if p.Range == PAGE {
		query.Skip((p.Page - 1) * p.Size).Limit(p.Size)
	}
	// 触发前置钩子
	if s, ok := schema.Origin.(IPreRestList); ok {
		s.PreRestList(c, query, p)
	}
	var list []kuu.H
	findErr := query.All(&list)
	if findErr != nil {
		handleError(findErr, c)
		return
	}
	query.Sort(p.Sort...)
	if list == nil {
		list = make([]kuu.H, 0)
	}
	// 构造返回
	data := kuu.H{
		"list":         list,
		"totalrecords": totalrecords,
	}
	if p.Range == PAGE {
		totalpages := int(math.Ceil(float64(totalrecords) / float64(p.Size)))
		data["totalpages"] = totalpages
		data["page"] = p.Page
		data["size"] = p.Size
	}
	if p.Sort != nil && len(p.Sort) > 0 {
		data["sort"] = p.Sort
	}
	if p.Project != nil {
		data["project"] = p.Project
	}
	if p.Cond != nil {
		data["cond"] = p.Cond
	}
	if p.Range != "" {
		data["range"] = p.Range
	}
	// 触发后置钩子
	if s, ok := schema.Origin.(IPostRestList); ok {
		s.PostRestList(c, &data)
	}
	c.JSON(http.StatusOK, kuu.StdOK(data))
}
