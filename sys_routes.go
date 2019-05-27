package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"path"
	"strings"
)

// OrgLoginRoute
var OrgLoginRoute = KuuRouteInfo{
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
		if data, err := ExecOrgLogin(sign, body.OrgID); err != nil {
			c.STDErrHold("组织登录失败").Data(err).Render()
		} else {
			c.STD(data)
		}
	},
}

// OrgListRoute
var OrgListRoute = KuuRouteInfo{
	Method: "GET",
	Path:   "/org/list",
	HandlerFunc: func(c *Context) {
		sign := c.SignInfo
		if data, err := GetOrgList(c.Context, sign.UID); err != nil {
			c.STDErrHold(c.L("获取组织列表失败")).Data(err).Render()
		} else {
			c.STD(data)
		}
	},
}

// OrgCurrentRoute
var OrgCurrentRoute = KuuRouteInfo{
	Method: "GET",
	Path:   "/org/current",
	HandlerFunc: func(c *Context) {
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
var UserRolesRoute = KuuRouteInfo{
	Method: "GET",
	Path:   "/user/roles",
	HandlerFunc: func(c *Context) {
		raw := c.Query("uid")
		if raw == "" {
			c.STDErr("用户ID不能为空")
			return
		}
		uid := ParseID(raw)
		if user, err := GetUserRoles(c.Context, uid); err != nil {
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
var UserMenusRoute = KuuRouteInfo{
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
var UploadRoute = KuuRouteInfo{
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
var AuthRoute = KuuRouteInfo{
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
var MetaRoute = KuuRouteInfo{
	Method: "GET",
	Path:   "/meta",
	HandlerFunc: func(c *Context) {
		c.STD(metadata)
	},
}
