package kuu

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func preflight() bool {
	var param Param
	DB().Where(&Param{Code: initCode, IsBuiltIn: true}).Find(&param)
	if param.ID != 0 {
		return true
	}
	return false
}

func createPresetDicts(tx *gorm.DB) {
	tx.Create(&Dict{
		Code:      "sys_menu_type",
		Name:      "菜单类型",
		IsBuiltIn: true,
		Values: []DictValue{
			{
				Label: "菜单",
				Value: "menu",
				Sort:  100,
			},
			{
				Label: "权限",
				Value: "permission",
				Sort:  200,
			},
		},
	})
	tx.Create(&Dict{
		Code:      "sys_data_range",
		Name:      "数据范围",
		IsBuiltIn: true,
		Values: []DictValue{
			{
				Label: "个人范围",
				Value: "personal",
				Sort:  100,
			},
			{
				Label: "当前组织范围",
				Value: "current",
				Sort:  200,
			},
			{
				Label: "当前及以下组织范围",
				Value: "current_following",
				Sort:  300,
			},
		},
	})
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
		root := User{
			Username:  "root",
			Name:      "预置用户",
			Password:  MD5("kuu"),
			IsBuiltIn: true,
		}
		tx.Create(&root)
		rootUser = &root
		// 初始化字典、菜单
		createPresetDicts(tx)
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

// GetSignContext
func GetSignContext(c *gin.Context) (sign *SignContext) {
	// 解析登录信息
	if v, exists := c.Get(SignContextKey); exists {
		sign = v.(*SignContext)
	} else {
		if v, err := DecodedContext(c); err == nil {
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
			return &data, errors.New(L(c, "未找到组织授权记录"))
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
		return &data, errors.New(L(c, "查询组织失败"))
	}
	return &data, nil
}

// GetUserRoles
func GetUserRoles(c *gin.Context, uid uint) (*User, error) {
	// 查询用户档案
	var user User
	if errs := DB().Where("id = ?", uid).Preload("RoleAssigns").First(&user).GetErrors(); len(errs) > 0 || user.ID == 0 {
		ERROR(errs)
		return &user, errors.New(L(c, "查询用户失败"))
	}
	// 过滤有效的角色分配
	var roleIDs []uint
	for _, assign := range user.RoleAssigns {
		if assign.ExpireUnix <= 0 || time.Now().Before(time.Unix(assign.ExpireUnix, 0)) {
			roleIDs = append(roleIDs, assign.RoleID)
		}
	}
	// 查询角色档案
	var (
		roles   []Role
		roleMap = make(map[uint]*Role)
	)
	if errs := DB().Where("id in (?)", roleIDs).Preload("OperationPrivileges").Preload("DataPrivileges").Find(&roles).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		return &user, errors.New(L(c, "查询角色失败"))
	}
	for _, role := range roles {
		roleMap[role.ID] = &role
	}
	// 重新赋值
	for index, assign := range user.RoleAssigns {
		assign.Role = roleMap[assign.RoleID]
		user.RoleAssigns[index] = assign
	}
	return &user, nil
}

// GetUserOrgs 查询用户组织
func GetUserOrgs(c *gin.Context, roles *[]Role) (*[]Org, error) {
	var orgs []Org
	if roles == nil {
		return &orgs, errors.New(L(c, "角色列表不存在"))
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
		return &orgs, errors.New(L(c, "查询角色失败"))
	}
	return &orgs, nil
}

// ExecOrgLogin
func ExecOrgLogin(c *gin.Context, sign *SignContext, orgID uint) (*Org, error) {
	var orgData Org
	if errs := DB().Where("id = ?", orgID).First(&orgData).GetErrors(); len(errs) > 0 || orgData.ID == 0 {
		ERROR(errs)
		return &orgData, errors.New(L(c, "组织不存在"))
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
		return &orgData, errors.New(L(c, "组织登录失败"))
	}
	// 缓存secret至redis
	key := RedisKeyBuilder(RedisOrgKey, signOrg.Token)
	value := Stringify(&signOrg)
	if _, err := RedisClient.Set(key, value, time.Second*time.Duration(ExpiresSeconds)).Result(); err != nil {
		ERROR(err)
	}
	c.SetCookie(OrgIDKey, strconv.Itoa(int(orgData.ID)), 0, "/", "", false, true)
	return &orgData, nil
}

func defaultLoginHandler(c *gin.Context) (jwt.MapClaims, uint, error) {
	body := struct {
		Username string
		Password string
	}{}
	// 解析请求参数
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		ERROR(err)
		return nil, 0, errors.New(L(c, "解析请求体失败"))
	}
	// 检测账号是否存在
	var user User
	if err := DB().Where(&User{Username: body.Username}).First(&user).Error; err != nil {
		ERROR(err)
		return nil, 0, errors.New(L(c, "用户不存在"))
	}
	// 检测账号是否有效
	if user.Disable {
		return nil, 0, errors.New(L(c, "该用户已被禁用"))
	}
	// 检测密码是否正确
	if !CompareHashAndPassword(user.Password, body.Password) {
		return nil, 0, errors.New(L(c, "账号密码不一致"))
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
	return payload, user.ID, nil
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
			MetaRoute,
			AuthRoute,
		},
		AfterImport: initSys,
	}
}
