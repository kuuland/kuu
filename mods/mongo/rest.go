package mongo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

var (
	defaultMessages = map[string]string{
		"request_error": "Request failed.",
	}
)

// ParseParams 解析请求上线文
func ParseParams(c *gin.Context) *Params {
	p := Params{}
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

// MountAll 挂载模型RESTful接口
func MountAll(k *kuu.Kuu, name string) {
	path := kuu.Join("/", strings.ToLower(name))
	k.POST(path, Create(k, name))
	k.DELETE(path, Remove(k, name))
	k.PUT(path, Update(k, name))
	k.GET(path, List(k, name))
	k.GET(kuu.Join(path, "/id/:id"), ID(k, name))
}

func setCreatedBy(c *gin.Context, docs []kuu.H) []kuu.H {
	var jwtData jwt.MapClaims
	if value, exists := c.Get("JWTDecoded"); exists && value != nil {
		jwtData = value.(jwt.MapClaims)
	}
	for _, item := range docs {
		if jwtData != nil && jwtData["_id"] != nil {
			createdBy := jwtData["_id"]
			switch createdBy.(type) {
			case string:
				createdBy = bson.ObjectIdHex(createdBy.(string))
			case bson.ObjectId:
				createdBy = createdBy.(bson.ObjectId)
			}
			item["CreatedBy"] = createdBy
		}
	}
	return docs
}

func setUpdatedBy(c *gin.Context, data kuu.H) kuu.H {
	var jwtData jwt.MapClaims
	if value, exists := c.Get("JWTDecoded"); exists && value != nil {
		jwtData = value.(jwt.MapClaims)
	}
	if data["UpdatedBy"] == nil && jwtData != nil && jwtData["_id"] != nil {
		updatedBy := jwtData["_id"]
		switch updatedBy.(type) {
		case string:
			updatedBy = bson.ObjectIdHex(updatedBy.(string))
		case bson.ObjectId:
			updatedBy = updatedBy.(bson.ObjectId)
		}
		data["UpdatedBy"] = updatedBy
	}
	return data
}
