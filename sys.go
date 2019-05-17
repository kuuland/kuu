package kuu

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/hoisie/mustache"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strconv"
	"strings"
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
	DB().Where(&Param{Code: initCode, IsBuiltIn: true}).Find(&param)
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
func GetUserRoles(c *gin.Context, uid uint) (*[]Role, *User, error) {
	var roles []Role
	// 查询用户档案
	var user User
	if errs := DB().Where("id = ?", uid).First(&user).GetErrors(); len(errs) > 0 || user.ID == 0 {
		ERROR(errs)
		return &roles, &user, errors.New(L(c, "Query user failed"))
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
		return &roles, &user, errors.New(L(c, "Query user roles failed"))
	}
	return &roles, &user, nil
}

// GetUserPermissions
func GetUserPermissions(c *gin.Context, uid uint, roles *[]Role) []string {
	if roles == nil {
		roles, _, _ = GetUserRoles(c, uid)
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
func GetUserOrgs(c *gin.Context, roles *[]Role) (*[]Org, error) {
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

// UpdateAuthRules
func UpdateAuthRules(tx *gorm.DB) {
	var (
		list   []User
		commit bool
	)
	if tx == nil {
		commit = true
		tx = DB().Begin()
	}
	tx.Find(&list)
	var rules []AuthRule
	if list != nil {
		for _, user := range list {
			if v := GenAuthRules(user.ID); v != nil && len(*v) > 0 {
				rules = append(rules, (*v)...)
			}
		}
		tx.Unscoped().Delete(&AuthRule{})
		tx.Exec(genRulesSQL(rules))
		if commit {
			if errs := tx.GetErrors(); len(errs) > 0 {
				ERROR(errs)
				if err := tx.Rollback().Error; err != nil {
					ERROR(err)
				}
			} else {
				if err := tx.Commit().Error; err != nil {
					ERROR(err)
				} else {
					INFO("Authorization rules have been updated")
				}
			}
		}
	}
}

func genRulesSQL(rules []AuthRule) string {
	var rows []string
	for _, r := range rules {
		row := mustache.Render("({{UID}},'{{Username}}',{{TargetOrgID}},'{{ObjectName}}',"+
			"'{{ReadableScope}}','{{WritableScope}}','{{ReadableOrgIDs}}','{{WritableOrgIDs}}','{{HitAssign}}','{{Permissions}}')", r)
		rows = append(rows, row)
	}
	sql := `"uid", "username", "target_org_id", "object_name", "readable_scope", "writable_scope", "readable_org_ids", "writable_org_ids", "hit_assign", "permissions"`
	sql = fmt.Sprintf(`INSERT INTO auth_rules (%s) VALUES %s`, sql, strings.Join(rows, ","))
	return sql
}

// GenAuthRules
func GenAuthRules(uid uint) *[]AuthRule {
	// 查询用户角色列表
	roles, user, _ := GetUserRoles(nil, uid)
	// 查询用户权限列表
	permissions := GetUserPermissions(nil, uid, roles)
	if roles == nil || permissions == nil {
		return nil
	}
	// 查询用户组织列表
	orgs, err := GetUserOrgs(nil, roles)
	if err != nil {
		ERROR(err)
		return nil
	}
	// 查询所有组织
	var totalOrgs []Org
	if err := DB().Select("id, full_path_pid").Find(&totalOrgs).Error; err != nil {
		ERROR(err)
		return nil
	}
	// 构建规则列表
	orgPrivilegesMap := getOrgPrivilegesMap(roles)
	var rules []AuthRule
	for _, org := range *orgs {
		//【直接授权】：直接针对组织的授权称为“直接授权”；
		//【间接授权】：在授权上级组织时选择了“当前及以下组织”而获得的称为“间接授权”。
		// 1.首先取直接授权
		// 2.若无直接授权，沿着组织树向上查询最近的一个“current_following”授权
		privilege := orgPrivilegesMap[org.ID]
		privilegeGetter := func(callback func(DataPrivileges) bool) {
			pids := strings.Split(org.FullPathPid, ",")
			for _, item := range pids {
				pid := ParseID(item)
				if pid == org.ID {
					continue
				}
				p := orgPrivilegesMap[pid]
				if p.OrgID != 0 && callback(p) {
					return
				}
			}
		}
		if privilege.OrgID == 0 {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllReadableRange == "current_following" || p.AllWritableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.AllReadableRange == "" {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllReadableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.AllWritableRange == "" {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllWritableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.OrgID == 0 {
			continue
		}
		authObjectsMaps := map[string]AuthObject{}
		if privilege.AuthObjects != nil {
			for _, authObject := range privilege.AuthObjects {
				authObjectsMaps[authObject.Name] = authObject
			}
		}
		metaArr := Metalist()
		for _, meta := range metaArr {
			authObject := authObjectsMaps[meta.Name]
			if authObject.Name == "" {
				authObject.Name = meta.Name
				authObject.ObjReadableRange = privilege.AllReadableRange
				authObject.ObjWritableRange = privilege.AllWritableRange
			}
			if authObject.ObjReadableRange == "" {
				authObject.ObjReadableRange = privilege.AllReadableRange
			}
			if authObject.ObjWritableRange == "" {
				authObject.ObjWritableRange = privilege.AllWritableRange
			}
			authObjectsMaps[meta.Name] = authObject
			rule := AuthRule{
				UID:           uid,
				Username:      user.Username,
				Name:          user.Name,
				TargetOrgID:   org.ID,
				ObjectName:    authObject.Name,
				ReadableScope: authObject.ObjReadableRange,
				WritableScope: authObject.ObjWritableRange,
				HitAssign:     authObject.ID,
				Permissions:   strings.Join(permissions, ","),
			}
			var (
				readableOrgIDs []string
				writableOrgIDs []string
			)
			switch rule.ReadableScope {
			case "current":
				readableOrgIDs = append(readableOrgIDs, string(rule.OrgID))
			case "current_following":
				for _, childOrg := range totalOrgs {
					if strings.HasPrefix(childOrg.FullPathPid, org.FullPathPid) {
						readableOrgIDs = append(readableOrgIDs, string(childOrg.ID))
					}
				}
			}
			switch rule.WritableScope {
			case "current":
				writableOrgIDs = append(writableOrgIDs, string(rule.OrgID))
			case "current_following":
				for _, childOrg := range totalOrgs {
					if strings.HasPrefix(childOrg.FullPathPid, org.FullPathPid) {
						writableOrgIDs = append(writableOrgIDs, string(childOrg.ID))
					}
				}
			}
			rule.ReadableOrgIDs = strings.Join(readableOrgIDs, ",")
			rule.WritableOrgIDs = strings.Join(writableOrgIDs, ",")
			rules = append(rules, rule)
		}
	}
	return &rules
}

func getOrgPrivilegesMap(roles *[]Role) (groups map[uint]DataPrivileges) {
	for _, role := range *roles {
		if role.DataPrivileges != nil {
			for _, privilege := range role.DataPrivileges {
				groups[privilege.OrgID] = privilege
			}
		}
	}
	return groups
}

// DefaultLoginHandler
func DefaultLoginHandler(c *gin.Context) (jwt.MapClaims, error) {
	body := struct {
		Username string
		Password string
	}{}
	// 解析请求参数
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		ERROR(err)
		return nil, errors.New(L(c, "Parsing body failed"))
	}
	// 检测账号是否存在
	var user User
	if err := DB().Where(&User{Username: body.Username}).First(&user).Error; err != nil {
		ERROR(err)
		return nil, errors.New(L(c, "User does not exist"))
	}
	// 检测账号是否有效
	if user.Disable {
		return nil, errors.New(L(c, "User has been disabled"))
	}
	// 检测密码是否正确
	if !CompareHashAndPassword(user.Password, body.Password) {
		return nil, errors.New(L(c, "Inconsistent password"))
	}
	payload := jwt.MapClaims{
		"UID":       user.ID,
		"Username":  user.Username,
		"Name":      user.Name,
		"Avatar":    user.Avatar,
		"Sex":       user.Sex,
		"Mobile":    user.Mobile,
		"Email":     user.Email,
		"IsBuiltIn": user.IsBuiltIn,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	}
	return payload, nil
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
			&AuthRule{},
			&Dict{},
			&DictValue{},
			&File{},
			&SignOrg{},
			&Param{},
			&Metadata{},
			&MetadataField{},
			&Route{},
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
