package kuu

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var valueCacheMap sync.Map

// OrgLoginRoute
var OrgLoginRoute = RouteInfo{
	Method: "POST",
	Path:   "/org/login",
	HandlerFunc: func(c *Context) {
		sign := c.SignInfo
		// 查询组织信息
		body := struct {
			OrgID uint `json:"org_id"`
		}{}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.OrgID == 0 {
			c.STDErr("解析请求体失败")
			return
		}
		c.IgnoreAuth()
		if data, err := ExecOrgLogin(sign, body.OrgID); err != nil {
			c.STDErr("组织登录失败", err)
		} else {
			c.STD(data)
		}
	},
}

// OrgListRoute
var OrgListRoute = RouteInfo{
	Method: "GET",
	Path:   "/org/list",
	HandlerFunc: func(c *Context) {
		c.IgnoreAuth()
		sign := c.SignInfo
		if data, err := GetOrgList(c.Context, sign.UID); err != nil {
			c.STDErr("获取组织列表失败", err)
		} else {
			c.STD(data)
		}
	},
}

// OrgCurrentRoute
var OrgCurrentRoute = RouteInfo{
	Method: "GET",
	Path:   "/org/current",
	HandlerFunc: func(c *Context) {
		c.IgnoreAuth()
		sign := c.SignInfo
		var signOrg SignOrg
		db := DB().Select("org_id").Order("created_at desc").Where(&SignOrg{UID: sign.UID, Token: sign.Token}).Preload("Org").First(&signOrg)
		if err := db.Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			ERROR(err)
			c.STDErr("未找到登录组织")
			return
		}
		var org Org
		DB().Where("id = ?", signOrg.OrgID).First(&org)
		c.STD(org)
	},
}

// UserRoleAssigns
var UserRoleAssigns = RouteInfo{
	Method: "GET",
	Path:   "/user/role_assigns/:uid",
	HandlerFunc: func(c *Context) {
		raw := c.Param("uid")
		if raw == "" {
			c.STDErr("用户ID不能为空")
			return
		}
		uid := ParseID(raw)
		if user, err := GetUserWithRoles(uid); err != nil {
			ERROR(err)
			c.STDErr(err.Error())
		} else {
			c.STD(user.RoleAssigns)
		}
	},
}

// UserMenusRoute
var UserMenusRoute = RouteInfo{
	Method: "GET",
	Path:   "/user/menus",
	HandlerFunc: func(c *Context) {
		c.SetRoutineCache(GLSUserMenusKey, true)
		var menus []Menu
		// 查询授权菜单
		if err := c.DB().Find(&menus).Error; err != nil {
			c.STDErr("菜单查询失败", err)
			return
		}
		// 补全父级菜单
		var total []Menu
		if err := c.IgnoreAuth().DB().Find(&total).Error; err != nil {
			c.STDErr("菜单查询失败", err)
			return
		}
		var (
			totalMap  = make(map[uint]Menu)
			existsMap = make(map[uint]bool)
			finded    = make(map[uint]bool)
		)
		for _, item := range total {
			totalMap[item.ID] = item
		}
		for _, item := range menus {
			existsMap[item.ID] = true
		}
		var fall func(result []Menu) []Menu
		fall = func(result []Menu) []Menu {
			recall := false
			for _, item := range result {
				if !finded[item.ID] {
					pitem := totalMap[item.Pid]
					if item.Pid != 0 && pitem.ID != 0 && !existsMap[pitem.ID] {
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
		c.STD(menus)
	},
}

// UploadRoute
var UploadRoute = RouteInfo{
	Method: "POST",
	Path:   "/upload",
	HandlerFunc: func(c *Context) {
		// 检查上传目录
		uploadDir := C().GetString("uploadDir")
		if uploadDir == "" {
			uploadDir = "assets"
		}
		EnsureDir(uploadDir)

		//	MD5
		file, _ := c.FormFile("file")
		src, err := file.Open()
		if err != nil {
			ERROR(err)
			c.STDErr(err.Error())
			return
		}
		defer func() {
			if err := src.Close(); err != nil {
				ERROR(err)
			}
		}()
		body, err := ioutil.ReadAll(src)
		md5 := fmt.Sprintf("%x", md5.Sum(body))

		//保存文件
		temps := strings.Split(file.Filename, ".")
		temps[0] = md5
		md5Name := strings.Join(temps, ".")
		dst := path.Join(uploadDir, md5Name)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ERROR(err)
			c.STDErr("保存上传文件失败")
			return
		}
		INFO(fmt.Sprintf("'%s' uploaded!", dst))

		class := c.PostForm("class")
		refidstr := c.PostForm("refid")
		refid := (uint)(0)
		if refidstr != "" {
			temp, err := strconv.ParseUint(refidstr, 10, 64)
			if err != nil {
				ERROR(err)
				c.STDErr(err.Error())
				return
			}
			refid = (uint)(temp)
		}
		f := new(File)
		f.Class = class
		f.RefID = refid
		f.UID = uuid.NewV4().String()
		f.Type = file.Header["Content-Type"][0]
		f.Size = file.Size
		f.Name = file.Filename
		f.Status = "done"
		f.URL = "/assets/" + md5Name
		f.Path = dst

		if errs := DB().Create(&f).GetErrors(); len(errs) > 0 {
			ERROR(errs)
			c.STDErr("保存上传文件失败")
		} else {
			c.STD(f)
		}
	},
}

// AuthRoute
var AuthRoute = RouteInfo{
	Method: "GET",
	Path:   "/auth",
	HandlerFunc: func(c *Context) {
		ps := c.Query("p")
		sp := strings.Split(ps, ",")

		if len(sp) == 0 {
			c.STDErr("权限编码不能为空")
		}

		ret := make(map[string]bool)
		for _, s := range sp {
			_, has := c.PrisDesc.Permissions[s]
			ret[s] = has
		}

		c.STD(ret)
	},
}

// MetaRoute
var MetaRoute = RouteInfo{
	Method: "GET",
	Path:   "/meta",
	HandlerFunc: func(c *Context) {
		json := c.Query("json")
		if json != "" {
			c.STD(metadataList)
		} else {
			var (
				hashKey = "meta"
				result  string
			)
			if v, ok := valueCacheMap.Load(hashKey); ok {
				result = v.(string)
			} else {
				var buffer bytes.Buffer
				for _, m := range metadataList {
					if len(m.Fields) > 0 {
						if m.DisplayName != "" {
							buffer.WriteString(fmt.Sprintf("%s(%s) {\n", m.Name, m.DisplayName))
						} else {
							buffer.WriteString(fmt.Sprintf("%s {\n", m.Name))
						}
						for index, field := range m.Fields {
							buffer.WriteString(fmt.Sprintf("\t%s(%s) %s", field.Code, field.Name, field.Type))
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
		}
	},
}

// EnumRoute
var EnumRoute = RouteInfo{
	Path:   "/enum",
	Method: "GET",
	HandlerFunc: func(c *Context) {
		json := c.Query("json")
		if json != "" {
			c.STD(enumList)
		} else {
			var (
				hashKey = "enum"
				result  string
			)
			if v, ok := valueCacheMap.Load(hashKey); ok {
				result = v.(string)
			} else {
				var buffer bytes.Buffer
				for _, desc := range enumList {
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
				result = buffer.String()
				valueCacheMap.Store(hashKey, result)
			}
			c.String(http.StatusOK, result)
		}
	},
}

// ModelDocsRoute
var ModelDocsRoute = RouteInfo{
	Method:       "GET",
	IgnorePrefix: true,
	Path:         "/model/docs",
	HandlerFunc: func(c *Context) {
		var (
			hashKeyYAML = "model_docs_yaml"
			hashKeyJSON = "model_docs_json"
		)
		// 取缓存
		if c.Query("yaml") != "" {
			if v, ok := valueCacheMap.Load(hashKeyYAML); ok {
				c.String(http.StatusOK, v.(string))
				return
			}
		} else {
			if v, ok := valueCacheMap.Load(hashKeyJSON); ok {
				c.String(http.StatusOK, v.(string))
				return
			}
		}
		// 重新生成
		var validMeta []*Metadata
		for _, m := range metadataList {
			if !m.RestDesc.IsValid() || len(m.Fields) == 0 {
				continue
			}
			validMeta = append(validMeta, m)
		}

		name := C().DefaultGetString("name", "Kuu")
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
					key := strings.ToLower(fmt.Sprintf("/%s", m.Name))
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
							if f.Enum != "" && enumMap[f.Enum] != nil {
								for value, _ := range enumMap[f.Enum].Values {
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
		if c.Query("yaml") != "" {
			valueCacheMap.Store(hashKeyYAML, yml)
			c.String(http.StatusOK, yml)
		} else {
			data, e := yaml.YAMLToJSON([]byte(yml))
			if e != nil {
				c.STDErr(e.Error())
				return
			}
			json := string(data)
			valueCacheMap.Store(hashKeyJSON, json)
			c.String(http.StatusOK, json)
		}
	},
}
