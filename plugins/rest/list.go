package rest

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

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
		totalrecords, _ := query.Count()
		if p.project != nil {
			query.Select(p.project)
		}
		if p._range == PAGE {
			query.Skip((p.page - 1) * p.size).Limit(p.size)
		}
		var list []kuu.H
		query.All(&list)
		query.Sort(p.sort...)
		if list == nil {
			list = make([]kuu.H, 0)
		}
		// 返回结果
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

type params struct {
	page    int
	size    int
	_range  string
	sort    []string
	project map[string]int
	cond    map[string]interface{}
}

func parseParams(c *gin.Context) *params {
	p := params{}
	// 处理page参数
	if c.Query("page") != "" {
		n, err := strconv.Atoi(c.Query("page"))
		if err == nil {
			p.page = n
		}
	}
	// 处理size参数
	if c.Query("size") != "" {
		n, err := strconv.Atoi(c.Query("size"))
		if err == nil {
			p.size = n
		}
	}
	// 处理range参数
	p._range = strings.ToUpper(c.Query("range"))
	if p._range != ALL && p._range != PAGE {
		p._range = PAGE
	}
	if p._range == PAGE {
		// PAGE模式下赋分页默认值
		if p.page == 0 {
			p.page = 1
		}
		if p.size == 0 {
			p.size = 20
		}
	}
	// 处理sort参数
	if c.Query("sort") != "" {
		p.sort = strings.Split(c.Query("sort"), ",")
	}
	// 处理project参数
	if c.Query("project") != "" {
		p.project = make(map[string]int)
		project := strings.Split(c.Query("project"), ",")
		for _, item := range project {
			var (
				key   string
				value int
			)
			exclude := false
			if strings.HasPrefix(item, "-") {
				exclude = true
			}
			if exclude == true {
				key = item[1:len(item)]
				value = 0
			} else {
				key = item
				value = 1
			}
			p.project[key] = value
		}
	}
	// 处理cond参数
	if c.Query("cond") != "" {
		cond := make(map[string]interface{})
		json.Unmarshal([]byte(c.Query("cond")), &cond)
		p.cond = cond
		// 避免服务端注入攻击
		if p.cond["$where"] != nil {
			delete(p.cond, "$where")
		}
	}
	return &p
}
