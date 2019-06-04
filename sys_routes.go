package kuu

import (
	"bytes"
	"fmt"
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
		for _, m := range metadata {
			fmt.Println(m)
		}
		name := C().DefaultGetString("name", "Kuu")
		doc := Doc{
			Openapi: "3.0.1",
			Info: DocInfo{
				Title:       name,
				Description: fmt.Sprintf("%s 模型RESTful接口文档", name),
				Version:     "1.0.0",
				Contact: DocInfoContact{
					Email: "yinfxs@dexdev.me",
				},
			},
			Servers: []DocServer{
				{
					Url: "https://humansa.hofo.co",
				},
			},
			Tags: []DocTag{
				{
					Name:        "Member",
					Description: "会员资讯",
				},
				{
					Name:        "Address",
					Description: "会员地址",
				},
			},
			Paths: map[string]DocPathItem{
				"/member": {
					"post": DocPathDesc{
						Tags:        []string{"Member"},
						Summary:     "新增会员资讯",
						OperationID: "createMember",
						RequestBody: DocPathRequestBody{
							Required:    true,
							Description: "新增会员资讯信息",
							Content: map[string]DocPathContentItem{
								"application/json": {
									Schema: DocPathSchema{
										Ref: "#/components/schemas/Member",
									},
								},
							},
						},
						Responses: map[int]DocPathResponse{
							200: {
								Description: "接口调用成功",
								Content: map[string]DocPathContentItem{
									"application/json": {
										Schema: DocPathSchema{
											Ref: "#/components/schemas/Member",
										},
									},
								},
							},
						},
					},
				},
			},
			Components: DocComponent{
				Schemas: func() (schemas map[string]DocComponentSchema) {
					schemas = make(map[string]DocComponentSchema)
					for _, m := range metadata {
						props := make(map[string]DocSchemaProperty)
						for _, f := range m.Fields {
							prop := DocSchemaProperty{}
							if f.IsRef {
								prop.Ref = fmt.Sprintf("#/components/schemas/%s", f.Type)
							} else {
								prop.Type = f.Type
							}
							props[f.Name] = prop
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
		yaml := doc.Marshal()
		c.String(http.StatusOK, yaml)
	},
}
