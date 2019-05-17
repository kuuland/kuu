package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	uuid "github.com/satori/go.uuid"
	"path"
)

// OrgLoginRoute
var OrgLoginRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/org/login",
	HandlerFunc: func(c *gin.Context) {
		// 解析登录信息
		sign := ensureLogged(c)
		if sign == nil {
			return
		}
		// 查询组织信息
		body := struct {
			OrgID uint `json:"org_id"`
		}{}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.OrgID == 0 {
			STDErr(c, L(c, "Parsing body failed"))
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
		// 解析登录信息
		sign := ensureLogged(c)
		if sign == nil {
			return
		}
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
		sign := ensureLogged(c)
		if sign == nil {
			return
		}
		var signOrg SignOrg
		if errs := DB().Where(&SignOrg{UID: sign.UID, Token: sign.Token}).Preload("Org").First(&signOrg).GetErrors(); len(errs) > 0 {
			ERROR(errs)
			STDErr(c, L(c, "Query login organization failed"))
			return
		}
		STD(c, signOrg.Org)
	},
}

// UserRolesRoute
var UserRolesRoute = gin.RouteInfo{
	Method: "GET",
	Path:   "/user/roles",
	HandlerFunc: func(c *gin.Context) {
		sign := ensureLogged(c)
		if sign == nil {
			return
		}
		if roles, _, err := GetUserRoles(c, sign.UID); err != nil {
			ERROR(err)
			STDErr(c, err.Error())
		} else {
			STD(c, roles)
		}
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