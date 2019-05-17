package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const initCode = "sys:init"

var (
	rootUser *User
	OrgIDKey = "OrgID"
)

// RootUID
func RootUID() uint {
	if rootUser != nil {
		return rootUser.ID
	}
	return 0
}

// RootUser
func RootUser() *User {
	return rootUser
}

func getRootUser() *User {
	var root User
	if errs := DB().Model(&User{Username: "root"}).First(&root).GetErrors(); len(errs) > 0 {
		ERROR(errs)
	}
	return &root
}

func createRootUser(tx *gorm.DB) *User {
	root := User{
		Username:  "root",
		Name:      "预置用户",
		Password:  MD5("kuu"),
		IsBuiltIn: true,
	}
	if errs := DB().Create(&root).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		PANIC("create root user failed")
	}
	return &root
}

func preflight() bool {
	var param Param
	if errs := DB().Where(&Param{Code: initCode, IsBuiltIn: true}).Find(&param).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		PANIC("Query init parameter failed")
	}
	if param.ID != 0 {
		return true
	}
	return false
}

func createPresetMenus(tx *gorm.DB) {
	rootMenu := Menu{
		Name:      "主导航菜单",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&rootMenu)
	sysMenu := Menu{
		Pid:       rootMenu.ID,
		Name:      "系统管理",
		Icon:      "setting",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&sysMenu)
	orgMenu := Menu{
		Pid:       sysMenu.ID,
		Name:      "组织管理",
		Icon:      "appstore",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&orgMenu)
	userMenu := Menu{
		Pid:       orgMenu.ID,
		Name:      "用户管理",
		Icon:      "user",
		URI:       "/sys/user",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&userMenu)
	sysOrgMenu := Menu{
		Pid:       orgMenu.ID,
		Name:      "组织机构",
		Icon:      "cluster",
		URI:       "/sys/org",
		Sort:      200,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&sysOrgMenu)
	permissionMenu := Menu{
		Pid:       sysMenu.ID,
		Name:      "权限管理",
		Icon:      "dropbox",
		Sort:      200,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&permissionMenu)
	roleMenu := Menu{
		Pid:       permissionMenu.ID,
		Name:      "角色管理",
		Icon:      "team",
		URI:       "/sys/role",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&roleMenu)
	settingMenu := Menu{
		Pid:       sysMenu.ID,
		Name:      "系统设置",
		Icon:      "tool",
		Sort:      300,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&settingMenu)
	menuMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "菜单管理",
		Icon:      "bars",
		URI:       "/sys/menu",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&menuMenu)
	paramMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "参数管理",
		Icon:      "profile",
		URI:       "/sys/param",
		Sort:      200,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&paramMenu)
	dictMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "字典管理",
		Icon:      "build",
		URI:       "/sys/dict",
		Sort:      300,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&dictMenu)
	auditMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "审计日志",
		Icon:      "book",
		URI:       "/sys/audit",
		Sort:      400,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&auditMenu)
	fileMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "文件",
		Icon:      "file",
		URI:       "/sys/file",
		Sort:      500,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&fileMenu)
	i18nMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "国际化",
		Icon:      "global",
		URI:       "/sys/i18n",
		Sort:      600,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&i18nMenu)
	messageMenu := Menu{
		Pid:       settingMenu.ID,
		Name:      "消息",
		Icon:      "message",
		URI:       "/sys/message",
		Sort:      700,
		Type:      "menu",
		IsBuiltIn: true,
		Closeable: true,
	}
	tx.Create(&messageMenu)
}

func initSys() {
	if preflight() {
		rootUser = getRootUser()
	} else {
		tx := DB().Begin()
		// 初始化预置用户
		createRootUser(tx)
		root := User{
			Username:  "root",
			Name:      "预置用户",
			Password:  MD5("kuu"),
			IsBuiltIn: true,
		}
		tx.Create(&root)
		rootUser = &root
		// 初始化菜单
		createPresetMenus(tx)
		// 保存初始化标记
		param := Param{
			Code:      initCode,
			IsBuiltIn: true,
			Name:      "系统初始化标记",
			Value:     "ok",
		}
		param.CreatedByID = rootUser.ID
		param.UpdatedByID = rootUser.ID
		tx.Create(&param)
		// 统一提交
		if errs := tx.GetErrors(); len(errs) > 0 {
			ERROR(errs)
			if err := tx.Rollback().Error; err != nil {
				PANIC("Init menu rollback failed: %s", err.Error())
			}
		} else {
			if err := tx.Commit().Error; err != nil {
				PANIC("Init menu commit failed: %s", err.Error())
			}
		}
	}
}

func ensureLogged(c *gin.Context) (sign *SignContext) {
	// 解析登录信息
	if v, exists := c.Get(SignContextKey); exists {
		sign = v.(*SignContext)
	} else {
		if v, err := DecodedContext(c); err != nil {
			STDErr(c, err.Error(), 555)
		} else {
			sign = v
		}
	}
	return
}

// GetOrgList
func GetOrgList(c *gin.Context, uid uint) (*[]Org, error) {
	var (
		data []Org
		db   *gorm.DB
	)
	if uid == RootUID() {
		db = DB().Find(&data)
	} else {
		var rules []AuthRule
		if errs := DB().Where(&AuthRule{UID: uid}).Find(&rules).GetErrors(); len(errs) > 0 {
			ERROR(errs)
			return &data, errors.New(L(c, "Query authorization organization failed"))
		}
		// 去重
		existMap := make(map[uint]bool, 0)
		orgIDs := make([]uint, 0)
		if rules != nil {
			for _, rule := range rules {
				if existMap[rule.OrgID] {
					continue
				}
				existMap[rule.OrgID] = true
				orgIDs = append(orgIDs, rule.OrgID)
			}
		}
		db = DB().Where("id in (?)", orgIDs).Find(&data)
	}
	if errs := db.GetErrors(); len(errs) > 0 {
		return &data, errors.New(L(c, "Query organizations failed"))
	}
	return &data, nil
}

// GetUserRoles
func GetUserRoles(c *gin.Context, uid uint) (*[]Role, error) {
	var roles []Role
	// 查询用户档案
	var user User
	if errs := DB().Where("id = ?", uid).First(&user).GetErrors(); len(errs) > 0 || user.ID == 0 {
		ERROR(errs)
		return &roles, errors.New(L(c, "Query user failed"))
	}
	// 过滤有效的角色分配
	var roleIDs []uint
	if user.RoleAssigns != nil {
		for _, assign := range user.RoleAssigns {
			if assign.ExpiryUnix <= 0 || time.Now().Before(time.Unix(assign.ExpiryUnix, 0)) {
				roleIDs = append(roleIDs, assign.RoleID)
			}
		}
	}
	// 查询角色档案
	if errs := DB().Where("id in (?)", roleIDs).Find(&roles).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		return &roles, errors.New(L(c, "Query user roles failed"))
	}
	return &roles, nil
}

// GetUserPermissions
func GetUserPermissions(c *gin.Context, uid uint, roles *[]Role) []string {
	if roles == nil {
		roles, _ = GetUserRoles(c, uid)
	}
	permissions := []string{}
	if roles != nil {
		for _, role := range *roles {
			if role.OperationPrivileges == nil || len(role.OperationPrivileges) == 0 {
				continue
			}
			for _, privilege := range role.OperationPrivileges {
				permissions = append(permissions, privilege.Permission)
			}
		}
	}
	return permissions
}

// GetUserOrgs 查询用户组织
func GetUserOrgs(c *gin.Context, uid uint, roles *[]Role) (*[]Org, error) {
	var orgs []Org
	if roles == nil {
		return &orgs, errors.New(L(c, "Role list is required"))
	}
	// 提取组织ID
	var orgIDs []uint
	for _, role := range *roles {
		if role.DataPrivileges != nil {
			for _, item := range role.DataPrivileges {
				if item.OrgID != 0 {
					orgIDs = append(orgIDs, item.OrgID)
				}
			}
		}
	}
	// 查询组织列表
	if errs := DB().Where("id in (?)", orgIDs).Find(&orgs).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		return &orgs, errors.New(L(c, "Query user roles failed"))
	}
	return &orgs, nil
}

// ExecOrgLogin
func ExecOrgLogin(c *gin.Context, sign *SignContext, orgID uint) (*Org, error) {
	var orgData Org
	if errs := DB().Where("id = ?", orgID).First(&orgData).GetErrors(); len(errs) > 0 || orgData.ID == 0 {
		ERROR(errs)
		return &orgData, errors.New(L(c, "Organization does not exist"))
	}
	// 新增登入记录
	signOrg := SignOrg{
		UID:   sign.UID,
		Token: sign.Token,
	}
	signOrg.OrgID = orgData.ID
	signOrg.CreatedByID = sign.UID
	signOrg.UpdatedByID = sign.UID
	if errs := DB().Create(&signOrg).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		return &orgData, errors.New(L(c, "Organization login failed"))
	}
	c.SetCookie(OrgIDKey, string(orgData.ID), 0, "/", "", false, true)
	return &orgData, nil
}

// ParseOrgID
func ParseOrgID(c *gin.Context) (orgID uint) {
	// querystring > header > cookie
	var id string
	id = c.Query(OrgIDKey)
	if id == "" {
		id = c.GetHeader(OrgIDKey)
	}
	if id == "" {
		id, _ = c.Cookie(OrgIDKey)
	}
	if v, err := strconv.ParseUint(id, 10, 0); err != nil {
		ERROR(err)
	} else {
		orgID = uint(v)
	}
	return
}

// Sys
func Sys() *Mod {
	return &Mod{
		Models: []interface{}{
			&User{},
			&Org{},
			&RoleAssign{},
			&Role{},
			&OperationPrivileges{},
			&DataPrivileges{},
			&AuthObject{},
			&Menu{},
			&Audit{},
			&AuthRule{},
			&Dict{},
			&DictValue{},
			&File{},
			&SignOrg{},
			&Param{},
			&Message{},
		},
		Middleware: gin.HandlersChain{
			OrgMiddleware,
		},
		Routes: gin.RoutesInfo{
			OrgLoginRoute,
			OrgListRoute,
			OrgCurrentRoute,
			UserRolesRoute,
			UploadRoute,
		},
		AfterImport: initSys,
	}
}
