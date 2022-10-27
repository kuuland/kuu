package kuu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var valueCacheMap sync.Map

// OrgLoginableRoute
var OrgLoginableRoute = RouteInfo{
	Name:   "查询可登录组织",
	Method: "GET",
	Path:   "/org/loginable",
	IntlMessages: map[string]string{
		"org_query_failed": "Query organization failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		data, err := GetLoginableOrgs(c, c.SignInfo.UID)
		if err != nil {
			return c.STDErr(err, "org_query_failed")
		}
		return c.STD(data)
	},
}

// OrgSwitchRoute
var OrgSwitchRoute = RouteInfo{
	Name:   "切换当前登录组织",
	Method: "POST",
	Path:   "/org/switch",
	IntlMessages: map[string]string{
		"org_switch_failed": "Switching organization failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var body struct {
			ActOrgID uint
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "org_switch_failed")
		}

		err := c.IgnoreAuth().DB().
			Model(&User{ID: c.SignInfo.UID}).
			Update(User{ActOrgID: body.ActOrgID}).Error

		if err != nil {
			return c.STDErr(err, "org_switch_failed")
		}
		return c.STDOK()
	},
}

// UserRoleAssigns
var UserRoleAssigns = RouteInfo{
	Name:   "查询用户已分配角色",
	Method: "GET",
	Path:   "/user/role_assigns/:uid",
	IntlMessages: map[string]string{
		"role_assigns_failed": "User roles query failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		raw := c.Param("uid")
		if raw == "" {
			return c.STDErr(errors.New("UID is required"), "role_assigns_failed")
		}
		uid := ParseID(raw)
		user, err := GetUserWithRoles(uid)
		if err != nil {
			return c.STDErr(err, "role_assigns_failed")
		}
		return c.STD(user.RoleAssigns)
	},
}

type MenuList []Menu

func (ml MenuList) Len() int {
	return len(ml)
}

func (ml MenuList) Less(i, j int) bool {
	return ml[i].Sort.Int64 < ml[j].Sort.Int64
}
func (ml MenuList) Swap(i, j int) {
	tmp := ml[i]
	ml[i] = ml[j]
	ml[j] = tmp
}

// UserMenusRoute
var UserMenusRoute = RouteInfo{
	Name:   "查询用户菜单",
	Method: "GET",
	Path:   "/user/menus",
	IntlMessages: map[string]string{
		"user_menus_failed": "User menus query failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var menus MenuList
		// 查询授权菜单
		if err := c.DB().Find(&menus).Error; err != nil {
			return c.STDErr(err, "user_menus_failed")
		}
		// 补全父级菜单
		var total MenuList
		if err := c.IgnoreAuth().DB().Find(&total).Error; err != nil {
			return c.STDErr(err, "user_menus_failed")
		}
		var (
			codeMap   = make(map[string]Menu)
			existsMap = make(map[uint]bool)
			finded    = make(map[uint]bool)
		)
		for _, item := range total {
			codeMap[item.Code] = item
		}
		for _, item := range menus {
			existsMap[item.ID] = true
		}
		var fall func(result MenuList) MenuList
		fall = func(result MenuList) MenuList {
			recall := false
			for _, item := range result {
				if !finded[item.ID] {
					pitem := codeMap[item.ParentCode.String]
					if item.ParentCode.String != "" && pitem.ID != 0 && !existsMap[pitem.ID] {
						result = append(result, pitem)
						recall = true
						existsMap[pitem.ID] = true
					}
					finded[item.ID] = true
				}
			}
			if recall {
				return fall(result)
			}
			return result
		}
		menus = fall(menus)
		if strings.ToLower(c.DefaultQuery("default", "true")) == "false" {
			var filtered MenuList
			for _, item := range menus {
				if item.Code != "default" {
					filtered = append(filtered, item)
				}
			}
			menus = filtered
		}
		sort.Sort(menus)
		return c.STD(menus)
	},
}

func getFileExtraData(c *Context) (*File, error) {
	class := c.PostForm("Class")
	ownerID := (uint)(0)
	if v := c.PostForm("OwnerID"); v != "" {
		vv, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		ownerID = (uint)(vv)
	}
	ownerType := c.PostForm("OwnerType")
	return &File{Class: class, OwnerID: ownerID, OwnerType: ownerType}, nil
}

// UploadRoute
var UploadRoute = RouteInfo{
	Name:   "默认文件上传接口",
	Method: "POST",
	Path:   "/upload",
	IntlMessages: map[string]string{
		"upload_failed": "Upload file failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var (
			save2db = true
		)
		if v, ok := c.GetPostForm("save2db"); ok {
			if b, err := strconv.ParseBool(v); err == nil {
				save2db = b
			}
		}
		extra, err := getFileExtraData(c)
		if err != nil {
			return c.STDErr(err, "upload_failed")
		}
		fh, err := c.FormFile("file")
		if err != nil {
			return c.STDErr(err, "upload_failed")
		}
		file, err := SaveUploadedFile(fh, save2db, extra)
		if err != nil {
			return c.STDErr(err, "upload_failed")
		}
		return c.STD(file)
	},
}

// AuthRoute
var AuthRoute = RouteInfo{
	Name:   "操作权限鉴权接口",
	Method: "GET",
	Path:   "/auth",
	IntlMessages: map[string]string{
		"auth_failed": "Authentication failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		ps := c.Query("p")
		split := strings.Split(ps, ",")

		if len(split) == 0 {
			return c.STDErr(errors.New("param 'p' is required"), "auth_failed")
		}

		ret := make(map[string]bool)
		for _, s := range split {
			_, has := c.PrisDesc.PermissionMap[s]
			ret[s] = has
		}

		return c.STD(ret)
	},
}

// MetaRoute
var MetaRoute = RouteInfo{
	Name:   "查询元数据列表",
	Method: "GET",
	Path:   "/meta",
	HandlerFunc: func(c *Context) *STDReply {
		json := c.Query("json")
		name := c.Query("name")
		mod := c.Query("mod")

		var list []*Metadata
		if name != "" {
			for _, name := range strings.Split(name, ",") {
				if v, ok := metadataMap[name]; ok && v != nil {
					list = append(list, v)
				}
			}
		} else if mod != "" {
			for _, item := range strings.Split(mod, ",") {
				for _, meta := range metadataList {
					if meta.ModCode == item {
						list = append(list, meta)
					}
				}
			}
		} else {
			list = metadataList
		}
		if json != "" {
			return c.STD(list)
		} else {
			var (
				hashKey = fmt.Sprintf("meta_%s_%s", name, mod)
				result  string
			)
			if v, ok := valueCacheMap.Load(hashKey); ok {
				result = v.(string)
			} else {
				var buffer bytes.Buffer
				for _, m := range list {
					if len(m.Fields) > 0 {
						if m.DisplayName != "" {
							buffer.WriteString(fmt.Sprintf("%s(%s) {\n", m.Name, m.DisplayName))
						} else {
							buffer.WriteString(fmt.Sprintf("%s {\n", m.Name))
						}
						for index, field := range m.Fields {
							if field.Enum != "" {
								buffer.WriteString(fmt.Sprintf("\t%s %s ENUM(%s)", field.Code, field.Name, field.Enum))
							} else {
								buffer.WriteString(fmt.Sprintf("\t%s %s %s", field.Code, field.Name, field.Type))
							}

							if index != len(m.Fields)-1 {
								buffer.WriteString("\n")
							}
						}
						buffer.WriteString(fmt.Sprintf("\n}\n\n"))
					}
				}
				result = buffer.String()
				valueCacheMap.Store(hashKey, result)
			}
			c.String(http.StatusOK, result)
			return nil
		}
	},
}

// EnumRoute
var EnumRoute = RouteInfo{
	Name:   "查询枚举列表",
	Path:   "/enum",
	Method: "GET",
	HandlerFunc: func(c *Context) *STDReply {
		json := c.Query("json")
		name := c.Query("name")

		em := EnumMap()
		var list []*EnumDesc
		if name != "" {
			for _, name := range strings.Split(name, ",") {
				if v, ok := em[name]; ok && v != nil {
					list = append(list, v)
				}
			}
		} else {
			list = EnumList()
		}
		if json != "" {
			return c.STD(list)
		} else {
			var buffer bytes.Buffer
			for _, desc := range list {
				if desc.ClassName != "" {
					buffer.WriteString(fmt.Sprintf("%s(%s) {\n", desc.ClassCode, desc.ClassName))
				} else {
					buffer.WriteString(fmt.Sprintf("%s {\n", desc.ClassCode))
				}
				index := 0
				for value, label := range desc.Values {
					if len(label) < 20 {
						for i := 0; i < 20-len(label); i++ {
							label += " "
						}
					}
					buffer.WriteString(fmt.Sprintf("\t%s\t%v(%s)", label, value, reflect.ValueOf(value).Type().Kind().String()))
					if index != len(desc.Values)-1 {
						buffer.WriteString("\n")
					}
					index++
				}
				buffer.WriteString(fmt.Sprintf("\n}\n\n"))
			}
			c.String(http.StatusOK, buffer.String())
			return nil
		}
	},
}

// DataDictRoute
var DataDictRoute = RouteInfo{
	Name:   "查询数据字典",
	Method: "GET",
	Path:   "/datadict",
	HandlerFunc: func(c *Context) *STDReply {
		modCode := c.Query("modCode")
		var buff strings.Builder
		buff.WriteString(fmt.Sprintf("# %s数据字典\n\n", C().GetString("name")))
		var modname string
		bookmap := map[bool]string{true: "是", false: "否"}
		m := DefaultCache.HGetAll(BuildKey("datadict"))
		for _, item := range m {
			var meta Metadata
			err := JSONParse(item, &meta)
			if err != nil {
				return c.STDErr(err)
			}
			if meta.ModCode == "" {
				continue
			}
			if modCode != "" && meta.ModCode != modCode {
				continue
			}
			if modname != meta.ModCode {
				modname = meta.ModCode
				buff.WriteString(fmt.Sprintf("## %s\n\n", meta.ModCode))
			}
			buff.WriteString(fmt.Sprintf("### %s_%s %s\n\n", meta.ModCode, meta.NativeName, meta.DisplayName))
			buff.WriteString("|字段名|字段类型|是否可空|是否主键|注释|\n")
			buff.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
			for _, field := range meta.Fields {
				IsBland := bookmap[field.IsBland]
				IsPrimaryKey := bookmap[field.IsPrimaryKey]
				line := fmt.Sprintf("| %s | %s | %s | %s | %s |\n", field.NativeName, field.DBType, IsBland, IsPrimaryKey, field.Name)
				buff.WriteString(line)
			}
			buff.WriteString("\n\n")
		}
		c.String(http.StatusOK, buff.String())
		return nil
	},
}

// CaptchaRoute
var CaptchaRoute = RouteInfo{
	Name:   "查询验证码",
	Path:   "/captcha",
	Method: "GET",
	HandlerFunc: func(c *Context) *STDReply {
		var (
			user  = c.Query("user")
			valid bool
		)
		if user != "" {
			times := GetCacheInt(getFailedTimesKey(user))
			valid = failedTimesValid(times)
		}
		if valid == false {
			return c.STD(null.BoolFrom(valid))
		}
		// 生成验证码
		id, base64Str := GenerateCaptcha()
		c.SetCookie(CaptchaIDKey, id, ExpiresSeconds, "/", "", false, true)
		return c.STD(D{
			"id":        id,
			"base64Str": base64Str,
		})
	},
}

// ModelDocsRoute
var ModelDocsRoute = RouteInfo{
	Name:   "查询默认接口文档",
	Method: "GET",
	Path:   "/model/docs",
	IntlMessages: map[string]string{
		"model_docs_failed": "Model document query failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var (
			hashKeyYAML = "model_docs_yaml"
			hashKeyJSON = "model_docs_json"
		)

		json := c.Query("json") != ""
		// 取缓存
		if json {
			if v, ok := valueCacheMap.Load(hashKeyJSON); ok {
				c.String(http.StatusOK, v.(string))
				return nil
			}
		} else {
			if v, ok := valueCacheMap.Load(hashKeyYAML); ok {
				c.String(http.StatusOK, v.(string))
				return nil
			}
		}
		// 重新生成
		var validMeta []*Metadata
		for _, m := range metadataList {
			if m == nil || m.RestDesc == nil || !m.RestDesc.IsValid() || len(m.Fields) == 0 {
				continue
			}
			validMeta = append(validMeta, m)
		}

		name := GetAppName()
		doc := Doc{
			Openapi: "3.0.1",
			Info: DocInfo{
				Title: fmt.Sprintf("%s 模型默认接口文档", name),
				Description: "调用说明：\n" +
					"1. 本文档仅包含数据模型默认开放的增删改查RESTful接口\n" +
					"1. 接口请求/响应的Content-Type默认为application/json，UTF-8编码\n" +
					"1. 如未额外说明，接口响应格式默认为以下JSON格式：\n" +
					"\t- `code` - **业务状态码**，0表成功，非0表失败（错误码默认为-1，令牌失效为555），该值一定存在，请按照该值判断业务操作是否成功，`integer`\n" +
					"\t- `msg` - **提示信息**，表正常或异常情况下的提示信息，有值才存在，`string`\n" +
					"\t- `data` - **数据部分**，正常时返回请求数据，异常时返回错误详情，有值才存在，`类型视具体接口而定`\n" +
					"1. 日期格式为`2019-06-04T02:42:01.472Z`，js代码：`new Date().toISOString()`\n" +
					"1. 用户密码等信息统一为MD5加密后的32位小写字符串，npm推荐使用blueimp-md5" +
					"",
				Version: "1.0.0",
				Contact: DocInfoContact{
					Email: "yinfxs@dexdev.me",
				},
			},
			Servers: []DocServer{
				{Url: fmt.Sprintf("%s%s", c.Origin(), C().GetString("prefix")), Description: "默认服务器"},
			},
			Tags: func() (tags []DocTag) {
				tags = []DocTag{{Name: "辅助接口"}}
				for _, m := range validMeta {
					tags = append(tags, DocTag{
						Name:        m.Name,
						Description: m.DisplayName,
					})
				}
				return
			}(),
			Paths: func() (paths map[string]DocPathItems) {
				paths = map[string]DocPathItems{
					"/meta": {
						"get": {
							Tags:        []string{"辅助接口"},
							Summary:     "查询模型列表",
							OperationID: "meta",
							Responses: map[int]DocPathResponse{
								200: {
									Description: "查询模型列表成功",
									Content: map[string]DocPathContentItem{
										"text/plain": {
											Schema: DocPathSchema{Type: "string"},
										},
									},
								},
							},
						},
					},
					"/enum": {
						"get": {
							Tags:        []string{"辅助接口"},
							Summary:     "查询枚举列表",
							OperationID: "enum",
							Responses: map[int]DocPathResponse{
								200: {
									Description: "查询枚举列表成功",
									Content: map[string]DocPathContentItem{
										"text/plain": {
											Schema: DocPathSchema{Type: "string"},
										},
									},
								},
							},
						},
					},
					"/upload": {
						"post": {
							Tags:        []string{"辅助接口"},
							Summary:     "上传文件",
							OperationID: "upload",
							RequestBody: DocPathRequestBody{
								Content: map[string]DocPathContentItem{
									"multipart/form-data": {
										Schema: DocPathSchema{
											Type: "object",
											Properties: map[string]DocPathSchema{
												"file": {
													Type:        "string",
													Format:      "binary",
													Description: "文件",
												},
											},
										},
									},
								},
							},
							Responses: map[int]DocPathResponse{
								200: {
									Description: "上传成功",
									Content: map[string]DocPathContentItem{
										"application/json": {
											Schema: DocPathSchema{Type: "string"},
										},
									},
								},
							},
							Security: []DocPathItemSecurity{
								map[string][]string{
									"api_key": {},
								},
							},
						},
					},
					"/whitelist": {
						"get": {
							Tags:        []string{"辅助接口"},
							Summary:     "接口白名单",
							Description: "接口白名单是指`不需要任何令牌`，可直接访问的接口，请前往在线链接查看最新列表",
							OperationID: "whitelist",
							Responses: map[int]DocPathResponse{
								200: {
									Description: "查询接口白名单成功",
									Content: map[string]DocPathContentItem{
										"text/plain": {
											Schema: DocPathSchema{Type: "string"},
										},
									},
								},
							},
						},
					},
				}
				for _, m := range validMeta {
					key := strings.ToLower(path.Join(GetModPrefix(m.ModCode), fmt.Sprintf("/%s", m.Name)))
					items := make(DocPathItems)
					displayName := m.DisplayName
					if displayName == "" {
						displayName = m.Name
					}
					// 新增接口
					if m.RestDesc.Create {
						items["post"] = DocPathItem{
							Tags:        []string{m.Name},
							Summary:     fmt.Sprintf("新增%s", displayName),
							Description: "注意：\n1. 如需批量新增，请传递对象数组\n1. 当你请求体为对象格式时，返回数据也为对象格式\n1. 当你请求体为对象数组时，返回数据也为对象数组",
							OperationID: fmt.Sprintf("create%s", m.Name),
							RequestBody: DocPathRequestBody{
								Required:    true,
								Description: fmt.Sprintf("%s对象", displayName),
								Content: map[string]DocPathContentItem{
									"application/json": {
										Schema: DocPathSchema{
											Ref: fmt.Sprintf("#/components/schemas/%s", m.Name),
										},
									},
								},
							},
							Responses: map[int]DocPathResponse{
								200: {
									Description: fmt.Sprintf("新增%s成功", displayName),
									Content: map[string]DocPathContentItem{
										"application/json": {
											Schema: DocPathSchema{
												Ref: fmt.Sprintf("#/components/schemas/%s", m.Name),
											},
										},
									},
								},
							},
							Security: []DocPathItemSecurity{
								map[string][]string{
									"api_key": {},
								},
							},
						}
					}
					// 删除接口
					if m.RestDesc.Delete {
						items["delete"] = DocPathItem{
							Tags:        []string{m.Name},
							Summary:     fmt.Sprintf("删除%s", displayName),
							Description: "注意：\n如需批量删除，请指定multi=true",
							OperationID: fmt.Sprintf("delete%s", m.Name),
							Parameters: []DocPathParameter{
								{
									Name:        "cond",
									In:          "query",
									Required:    true,
									Description: "删除条件，JSON格式的字符串",
									Schema: DocPathSchema{
										Type: "string",
									},
								},
								{
									Name:        "multi",
									In:          "query",
									Description: "是否批量删除",
									Schema: DocPathSchema{
										Type: "boolean",
									},
								},
							},
							Responses: map[int]DocPathResponse{
								200: {
									Description: fmt.Sprintf("删除%s成功", displayName),
									Content: map[string]DocPathContentItem{
										"application/json": {
											Schema: DocPathSchema{
												Ref: fmt.Sprintf("#/components/schemas/%s", m.Name),
											},
										},
									},
								},
							},
							Security: []DocPathItemSecurity{
								map[string][]string{
									"api_key": {},
								},
							},
						}
					}
					// 修改接口
					if m.RestDesc.Update {
						items["put"] = DocPathItem{
							Tags:        []string{m.Name},
							Summary:     fmt.Sprintf("修改%s", displayName),
							Description: "注意：\n如需批量修改，请指定multi=true",
							OperationID: fmt.Sprintf("update%s", m.Name),
							RequestBody: DocPathRequestBody{
								Required:    true,
								Description: fmt.Sprintf("%s对象", displayName),
								Content: map[string]DocPathContentItem{
									"application/json": {
										Schema: DocPathSchema{
											Type: "object",
											Properties: map[string]DocPathSchema{
												"cond": {
													Ref:      fmt.Sprintf("#/components/schemas/%s", m.Name),
													Required: true,
												},
												"doc": {
													Ref:      fmt.Sprintf("#/components/schemas/%s", m.Name),
													Required: true,
												},
												"multi": {
													Type: "boolean",
												},
											},
										},
									},
								},
							},
							Responses: map[int]DocPathResponse{
								200: {
									Description: fmt.Sprintf("修改%s成功", displayName),
									Content: map[string]DocPathContentItem{
										"application/json": {
											Schema: DocPathSchema{
												Ref: fmt.Sprintf("#/components/schemas/%s", m.Name),
											},
										},
									},
								},
							},
							Security: []DocPathItemSecurity{
								map[string][]string{
									"api_key": {},
								},
							},
						}
					}
					// 查询接口
					if m.RestDesc.Query {
						items["get"] = DocPathItem{
							Tags:        []string{m.Name},
							Summary:     fmt.Sprintf("查询%s", displayName),
							OperationID: fmt.Sprintf("query%s", m.Name),
							Parameters: []DocPathParameter{
								{
									Name:        "range",
									In:          "query",
									Description: "查询数据范围，分页（PAGE）或全量（ALL）",
									Schema: DocPathSchema{
										Type: "string",
										Enum: []interface{}{
											"PAGE",
											"ALL",
										},
										Default: "PAGE",
									},
								},
								{
									Name:        "cond",
									In:          "query",
									Description: fmt.Sprintf("查询条件，%s对象的JSON字符串", displayName),
									Schema: DocPathSchema{
										Type: "string",
									},
								},
								{
									Name:        "sort",
									In:          "query",
									Description: "排序字段，多字段排序以英文逗号分隔，逆序以负号开头",
									Schema: DocPathSchema{
										Type: "string",
									},
								},
								{
									Name:        "project",
									In:          "query",
									Description: "查询字段，注意字段依然返回，只是不查询",
									Schema: DocPathSchema{
										Type: "string",
									},
								},
								{
									Name:        "page",
									In:          "query",
									Description: "当前页码（仅PAGE模式有效）",
									Schema: DocPathSchema{
										Type:    "integer",
										Default: 1,
									},
								},
								{
									Name:        "size",
									In:          "query",
									Description: "每页条数（仅PAGE模式有效）",
									Schema: DocPathSchema{
										Type:    "integer",
										Default: 30,
									},
								},
							},
							Responses: map[int]DocPathResponse{
								200: {
									Description: fmt.Sprintf("查询%s成功", displayName),
									Content: map[string]DocPathContentItem{
										"application/json": {
											Schema: DocPathSchema{
												Type: "object",
												Properties: map[string]DocPathSchema{
													"list": {
														Type: "array",
														Items: &DocPathSchema{
															Ref: fmt.Sprintf("#/components/schemas/%s", m.Name),
														},
													},
													"totalrecords": {
														Type:        "integer",
														Description: "当前查询条件下的总记录数",
													},
													"totalpages": {
														Type:        "integer",
														Description: "当前查询条件下的总页数（仅PAGE模式存在）",
													},
												},
											},
										},
									},
								},
							},
							Security: []DocPathItemSecurity{
								map[string][]string{
									"api_key": {},
								},
							},
						}
					}
					if len(items) > 0 {
						paths[key] = items
					}
				}
				return
			}(),
			Components: DocComponent{
				Schemas: func() (schemas map[string]DocComponentSchema) {
					schemas = make(map[string]DocComponentSchema)
					em := EnumMap()
					for _, m := range validMeta {
						props := make(map[string]DocSchemaProperty)
						for _, f := range m.Fields {
							prop := DocSchemaProperty{}
							if f.Name != "" {
								prop.Title = f.Name
							}
							if f.IsRef {
								if f.IsArray {
									prop.Type = "array"
									prop.Items = &DocSchemaProperty{
										Ref: fmt.Sprintf("#/components/schemas/%s", f.Type),
									}
								} else {
									prop.Ref = fmt.Sprintf("#/components/schemas/%s", f.Type)
								}
							} else {
								prop.Type = f.Type
							}
							if f.Enum != "" && em[f.Enum] != nil {
								for value, _ := range em[f.Enum].Values {
									prop.Enum = append(prop.Enum, value)
								}
							}
							props[f.Code] = prop
						}
						schemas[m.Name] = DocComponentSchema{
							Type:       "object",
							Properties: props,
						}
					}
					return
				}(),
				SecuritySchemes: map[string]DocSecurityScheme{
					"api_key": {
						Type: "apiKey",
						Name: "api_key",
						In:   "header",
					},
				},
			},
		}
		yml := doc.Marshal()
		if json {
			data, err := yaml.YAMLToJSON([]byte(yml))
			if err != nil {
				return c.STDErr(err, "model_docs_failed")
			}
			json := string(data)
			valueCacheMap.Store(hashKeyJSON, json)
			c.String(http.StatusOK, json)
		} else {
			valueCacheMap.Store(hashKeyYAML, yml)
			c.String(http.StatusOK, yml)
		}
		return nil
	},
}

// LangSwitchRoute
var LangSwitchRoute = RouteInfo{
	Name:   "切换用户语言环境",
	Method: "POST",
	Path:   "/lang/switch",
	IntlMessages: map[string]string{
		"lang_switch_failed": "Switching language failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var body struct {
			Lang string
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "lang_switch_failed")
		}

		err := c.IgnoreAuth().DB().
			Model(&User{ID: c.SignInfo.UID}).
			Update(User{Lang: body.Lang}).Error

		if err != nil {
			return c.STDErr(err, "lang_switch_failed")
		}
		return c.STDOK()
	},
}

// LoginAsRoute
var LoginAsRoute = RouteInfo{
	Name:   "以用户身份登录（该接口仅限root调用）",
	Method: "POST",
	Path:   "/login_as",
	IntlMessages: map[string]string{
		"login_as_unauthorized": "Unauthorized operation",
		"login_as_failed":       "Login failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		var body struct {
			UID uint
		}

		if c.SignInfo.UID != RootUID() {
			return c.STDErr(fmt.Errorf("unauthorized operation: uid=%v", c.SignInfo.UID), "login_as_unauthorized")
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			return c.STDErr(err, "login_as_failed")
		}

		var (
			secret SignSecret
			user   User
			db     = c.DB()
		)
		if err := db.Where(&SignSecret{UID: body.UID, Type: AdminSignType, Method: "LOGIN"}).Where(fmt.Sprintf("%s > ?", db.Dialect().Quote("exp")), time.Now().Unix()).Order("created_at desc").First(&secret).Error; err != nil {
			return c.STDErr(err, "login_as_failed")
		}
		if err := db.Where(fmt.Sprintf("%s = ?", db.Dialect().Quote("id")), secret.UID).First(&user).Error; err != nil {
			return c.STDErr(err, "login_as_failed")
		}
		c.SetCookie(LangKey, user.Lang, ExpiresSeconds, "/", "", false, true)
		c.SetCookie(TokenKey, secret.Token, ExpiresSeconds, "/", "", false, true)
		return c.STDOK()
	},
}

// LoginAsOutRoute
var LoginAsOutRoute = RouteInfo{
	Name:   "退出模拟登录（该接口仅限root调用）",
	Method: "DELETE",
	Path:   "/login_as",
	HandlerFunc: func(c *Context) *STDReply {
		c.SetCookie(TokenKey, c.SignInfo.Token, -1, "/", "", false, true)
		c.SetCookie(LangKey, "", -1, "/", "", false, true)
		return c.STDOK()
	},
}

// LoginAsUsersRoute
var LoginAsUsersRoute = RouteInfo{
	Name:   "查询可模拟登录的用户列表（该接口仅限root调用）",
	Method: "GET",
	Path:   "/login_as/users",
	IntlMessages: map[string]string{
		"login_as_unauthorized": "Unauthorized operation",
		"login_as_failed":       "Login failed",
	},
	HandlerFunc: func(c *Context) *STDReply {
		if c.SignInfo.UID != RootUID() {
			return c.STDErr(fmt.Errorf("unauthorized operation: uid=%v", c.SignInfo.UID), "login_as_unauthorized")
		}

		var (
			secrets []SignSecret
			uids    []uint
			users   []User
			db      = c.DB()
		)
		if err := db.Model(&SignSecret{}).Where(&SignSecret{Type: AdminSignType, Method: "LOGIN"}).Where(fmt.Sprintf("%s > ?", db.Dialect().Quote("exp")), time.Now().Unix()).Find(&secrets).Error; err != nil {
			return c.STDErr(err, "login_as_failed")
		}
		secretMap := make(map[uint]SignSecret)
		for _, item := range secrets {
			uids = append(uids, item.UID)
			secretMap[item.UID] = item
		}
		if err := db.Model(&User{}).Where(fmt.Sprintf("%s IN (?)", db.Dialect().Quote("id")), uids).Find(&users).Error; err != nil {
			return c.STDErr(err, "login_as_failed")
		}
		type record struct {
			ID       uint
			Name     string
			Username string
			Exp      int64
		}
		var records []record
		for _, item := range users {
			records = append(records, record{
				ID:       item.ID,
				Name:     item.Name,
				Username: item.Username,
				Exp:      secretMap[item.ID].Exp,
			})
		}
		return c.STD(records)
	},
}

// JobRunRoute
var JobRunRoute = RouteInfo{
	Name:   "触发定时任务立即运行接口",
	Method: http.MethodPost,
	Path:   "/job/run",
	HandlerFunc: func(c *Context) *STDReply {
		if err := RunJob(c.Query("code")); err != nil {
			return c.STDErr(err)
		}
		return c.STDOK()
	},
}
