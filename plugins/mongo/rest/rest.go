package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/db"
)

var (
	defaultMessages = map[string]string{
		"request_error": "Request failed.",
	}
	k      *kuu.Kuu
	name   string
	schema *kuu.Schema
)

// ParseParams 解析请求上线文
func ParseParams(c *gin.Context) *db.Params {
	p := db.Params{}
	if c.Param("id") != "" {
		p.ID = c.Param("id")
	}
	// 处理page参数
	if c.Query("page") != "" {
		n, err := strconv.Atoi(c.Query("page"))
		if err == nil {
			p.Page = n
		}
	}
	// 处理size参数
	if c.Query("size") != "" {
		n, err := strconv.Atoi(c.Query("size"))
		if err == nil {
			p.Size = n
		}
	}
	// 处理range参数
	p.Range = strings.ToUpper(c.Query("range"))
	if p.Range != db.ALL && p.Range != db.PAGE {
		p.Range = db.PAGE
	}
	if p.Range == db.PAGE {
		// PAGE模式下赋分页默认值
		if p.Page == 0 {
			p.Page = 1
		}
		if p.Size == 0 {
			p.Size = 20
		}
	}
	// 处理sort参数
	if c.Query("sort") != "" {
		p.Sort = strings.Split(c.Query("sort"), ",")
	}
	// 处理project参数
	if c.Query("project") != "" {
		p.Project = make(map[string]int)
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
				key = item[1 : len(item)-1]
				value = 0
			} else {
				key = item
				value = 1
			}
			p.Project[key] = value
		}
	}
	// 处理cond参数
	if c.Query("cond") != "" {
		var cond kuu.H
		json.Unmarshal([]byte(c.Query("cond")), &cond)
		p.Cond = cond
		// 避免服务端注入攻击
		if p.Cond["$where"] != nil {
			delete(p.Cond, "$where")
		}
	}
	return &p
}

func handleError(err error, c *gin.Context) {
	kuu.Error(err)
	c.JSON(http.StatusOK, kuu.StdError(kuu.SafeL(defaultMessages, c, "request_error")))
}

// Mount 挂载模型RESTful接口
func Mount(app *kuu.Kuu, n string) {
	k = app
	name = n
	schema = app.Schemas[name]
	path := kuu.Join("/", strings.ToLower(name))
	k.POST(path, create)
	k.DELETE(path, remove)
	k.PUT(path, update)
	k.GET(path, list)
	k.GET(kuu.Join(path, "/:id"), id)
}
