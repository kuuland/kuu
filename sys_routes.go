package kuu

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"path"
	"strings"
)

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
			c.STDErrHold("组织登录失败").Data(err).Render()
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
			c.STDErrHold(c.L("获取组织列表失败")).Data(err).Render()
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
		db := DB().Select("org_id").Where(&SignOrg{UID: sign.UID, Token: sign.Token}).Preload("Org").First(&signOrg)
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

// UserRolesRoute
var UserRolesRoute = RouteInfo{
	Method: "GET",
	Path:   "/user/roles",
	HandlerFunc: func(c *Context) {
		raw := c.Query("uid")
		if raw == "" {
			c.STDErr("用户ID不能为空")
			return
		}
		uid := ParseID(raw)
		if user, err := GetUserWithRoles(uid); err != nil {
			ERROR(err)
			c.STDErr(err.Error())
		} else {
			roles := make([]*Role, 0)
			for _, assign := range user.RoleAssigns {
				roles = append(roles, assign.Role)
			}
			c.STD(roles)
		}
	},
}

// UserMenusRoute
var UserMenusRoute = RouteInfo{
	Method: "GET",
	Path:   "/user/menus",
	HandlerFunc: func(c *Context) {
		desc := GetPrivilegesDesc(c.Context)
		var (
			menus []Menu
			db    = DB()
		)
		if desc.UID != RootUID() {
			db = db.Where("code in (?)", desc.Codes)
		}
		if err := db.Find(&menus).Error; err != nil {
			ERROR(err)
			c.STDErr("菜单查询失败")
			return
		}
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
		// 执行文件保存
		file, _ := c.FormFile("file")
		dst := path.Join(uploadDir, file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ERROR(err)
			c.STDErr("保存上传文件失败")
			return
		}
		INFO(fmt.Sprintf("'%s' uploaded!", dst))

		f := File{
			UID:    uuid.NewV4().String(),
			Type:   file.Header["Content-Type"][0],
			Size:   file.Size,
			Name:   file.Filename,
			Status: "done",
			URL:    "/assets/" + file.Filename,
			Path:   dst,
		}

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
			c.STD(metadata)
		} else {
			var buffer bytes.Buffer
			for _, m := range metadata {
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
			c.String(http.StatusOK, buffer.String())
		}
	},
}

// ModelDocsRoute
var ModelDocsRoute = RouteInfo{
	Method: "GET",
	Path:   "/model/docs",
	HandlerFunc: func(c *Context) {
		var validMeta []*Metadata
		for _, m := range metadata {
			if !m.RestDesc.IsValid() || len(m.Fields) == 0 {
				continue
			}
			validMeta = append(validMeta, m)
		}
		//if !m.RestDesc.IsValid() || len(m.Fields) == 0 {
		//	continue
		//}

		name := C().DefaultGetString("name", "Kuu")
		doc := Doc{
			Openapi: "3.0.1",
			Info: DocInfo{
				Title: fmt.Sprintf("%s 数据模型默认接口文档", name),
				Description: "调用说明：\n" +
					"1. 本文档仅包含数据模型默认开放的增删改查RESTful接口\n" +
					"1. 接口请求/响应的Content-Type默认为application/json，UTF-8编码\n" +
					"1. 如未额外说明，接口响应格式默认为以下JSON格式：\n" +
					"\t- `code` - **业务状态码**，0表成功，非0表失败（错误码默认为-1，令牌失效为555），该值一定存在，请按照该值判断业务操作是否成功，`integer`\n" +
					"\t- `msg` - **提示信息**，表正常或异常情况下的提示信息，有值才存在，`string`\n" +
					"\t- `data` - **数据部分**，正常时返回请求数据，异常时返回错误详情，有值才存在，`类型视具体接口而定`\n" +
					"1. 日期格式为\"2019-06-04T02:42:01.472Z\"，js代码：\"new Date().toISOString()\"\n" +
					"1. 用户密码等信息统一为MD5加密后的32位大写字符串，npm推荐使用blueimp-md5" +
					"",
				Version: "1.0.0",
				Contact: DocInfoContact{
					Email: "yinfxs@dexdev.me",
				},
			},
			Servers: []DocServer{
				{Url: c.Origin(), Description: "默认服务器"},
			},
			Tags: func() (tags []DocTag) {
				tags = make([]DocTag, 0)
				for _, m := range validMeta {
					tags = append(tags, DocTag{
						Name:        m.Name,
						Description: m.DisplayName,
					})
				}
				return
			}(),
			Paths: func() (paths map[string]DocPathItems) {
				paths = make(map[string]DocPathItems)
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
									"api_key": []string{},
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
									"api_key": []string{},
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
									"api_key": []string{},
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
										Ref:  fmt.Sprintf("#/components/schemas/%s", m.Name),
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
									"api_key": []string{},
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
			c.String(http.StatusOK, yml)
		} else {
			json, e := yaml.YAMLToJSON([]byte(yml))
			if e != nil {
				c.STDErr(e.Error())
				return
			}
			c.String(http.StatusOK, string(json))
		}
	},
}
