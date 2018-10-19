package rest

import (
	"encoding/json"
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

var (
	defaultMessages = map[string]string{
		"request_error": "Request failed.",
	}
	k      *kuu.Kuu
	name   string
	schema *kuu.Schema
	model  func(string) *mgo.Collection
)

// Params 从请求上下文中解析出的参数信息
type Params struct {
	Page    int
	Size    int
	Range   string
	Sort    []string
	Project map[string]int
	Cond    map[string]interface{}
}

// ParseParams 解析请求上线文
func ParseParams(c *gin.Context) *Params {
	p := Params{}
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
	if p.Range != ALL && p.Range != PAGE {
		p.Range = PAGE
	}
	if p.Range == PAGE {
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
				key = item[1:len(item)]
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
		cond := make(map[string]interface{})
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
func Mount(app *kuu.Kuu, n string, c func(string) *mgo.Collection) {
	k = app
	name = n
	model = c
	schema = app.Schemas[name]
	path := kuu.Join("/", strings.ToLower(name))
	k.POST(path, create)
	k.DELETE(path, remove)
	k.PUT(path, update)
	k.GET(path, list)
	k.GET(kuu.Join(path, "/:id"), id)
}
