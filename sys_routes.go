package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"path"
)

// OrgLoginRoute
var OrgLoginRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/org/login",
	HandlerFunc: func(c *gin.Context) {
		sign := GetSignContext(c)
		// 查询组织信息
		body := struct {
			OrgID uint `json:"org_id"`
		}{}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.OrgID == 0 {
			STDErr(c, L(c, "解析请求体失败"))
			return
		}
		if data, err := ExecOrgLogin(c, sign, body.OrgID); err != nil {
			ERROR(err)
			STDErr(c, err.Error())
		} else {
			STD(c, data)
		}
	},
}

// OrgListRoute
var OrgListRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/org/list",
	HandlerFunc: func(c *gin.Context) {
		sign := GetSignContext(c)
		if data, err := GetOrgList(c, sign.UID); err != nil {
			ERROR(err)
			STDErr(c, err.Error())
		} else {
			STD(c, data)
		}
	},
}

// OrgCurrentRoute
var OrgCurrentRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/org/current",
	HandlerFunc: func(c *gin.Context) {
		sign := GetSignContext(c)
		var signOrg SignOrg
		db := DB().Select("org_id").Where(&SignOrg{UID: sign.UID, Token: sign.Token}).Preload("Org").First(&signOrg)
		if err := db.Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			ERROR(err)
			STDErr(c, L(c, "未找到登录组织"))
			return
		}
		var org Org
		DB().Where("id = ?", signOrg.OrgID).First(&org)
		STD(c, org)
	},
}

// UserRolesRoute
var UserRolesRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/user/roles",
	HandlerFunc: func(c *gin.Context) {
		raw := c.Query("uid")
		if raw == "" {
			STDErr(c, L(c, "用户ID不能为空"))
			return
		}
		uid := ParseID(raw)
		if user, err := GetUserRoles(c, uid); err != nil {
			ERROR(err)
			STDErr(c, err.Error())
		} else {
			roles := make([]*Role, 0)
			for _, assign := range user.RoleAssigns {
				roles = append(roles, assign.Role)
			}
			STD(c, roles)
		}
	},
}

// UserMenusRoute
var UserMenusRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/user/menus",
	HandlerFunc: func(c *gin.Context) {
		desc := GetPrivilegesDesc(c)
		var (
			menus []Menu
			db    = DB()
		)
		if desc.UID != RootUID() {
			db = db.Where("code in (?)", desc.Codes)
		}
		if err := db.Find(&menus).Error; err != nil {
			ERROR(err)
			STDErr(c, L(c, "菜单查询失败"))
			return
		}
		STD(c, menus)
	},
}

// UploadRoute
var UploadRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/upload",
	HandlerFunc: func(c *gin.Context) {
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
			STDErr(c, "Saving uploaded file failed")
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
			STDErr(c, "Saving uploaded file failed")
		} else {
			STD(c, f)
		}
	},
}

// AuthRoute
var AuthRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/auth",
	HandlerFunc: func(c *gin.Context) {
		//permission := c.Query("p")
		//STD(c, metadata)
	},
}

// MetaRoute
var MetaRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/meta",
	HandlerFunc: func(c *gin.Context) {
		STD(c, metadata)
	},
}
