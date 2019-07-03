package kuu

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const initCode = "sys:init"

var rootUser *User

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
		Code:      "main",
		Name:      "主导航菜单",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&rootMenu)
	sysMenu := Menu{
		Code:      "sys",
		Pid:       rootMenu.ID,
		Name:      "系统管理",
		Icon:      "setting",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: true,
	}
	tx.Create(&sysMenu)
	orgMenu := Menu{
		Code:      "sys:omg",
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
		Code:      "sys:auth",
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
		Code:      "sys:settings",
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

func createMockData(tx *gorm.DB) {
	// 新增组织
	gzo := &Org{
		Code: "GZ",
		Name: "广州",
	}
	tx.Create(gzo)
	tho := &Org{
		Code: "TH",
		Name: "天河",
		Pid:  gzo.ID,
	}
	tx.Create(tho)
	yxo := &Org{
		Code: "YX",
		Name: "越秀",
		Pid:  gzo.ID,
	}
	tx.Create(yxo)
	lwo := &Org{
		Code: "LW",
		Name: "荔湾",
		Pid:  gzo.ID,
	}
	tx.Create(lwo)
	sho := &Org{
		Code: "SH",
		Name: "上海",
	}
	tx.Create(sho)
	// 新建角色
	gzr := &Role{
		Code: "gz_admin",
		Name: "广州管理员",
	}
	tx.Create(gzr)
	gzrOP := []OperationPrivileges{
		{
			RoleID:   gzr.ID,
			MenuCode: "sys:user",
		},
		{
			RoleID:   gzr.ID,
			MenuCode: "sys:org",
		},
		{
			RoleID:   gzr.ID,
			MenuCode: "sys:param",
		},
		{
			RoleID:   gzr.ID,
			MenuCode: "sys:role",
		},
	}
	for _, item := range gzrOP {
		tx.Create(&item)
	}
	gzrDP := []DataPrivileges{
		{
			RoleID:        gzr.ID,
			TargetOrgID:   gzo.ID,
			ReadableRange: "CURRENT_FOLLOWING",
			WritableRange: "CURRENT_FOLLOWING",
		},
		{
			RoleID:        gzr.ID,
			TargetOrgID:   tho.ID,
			ReadableRange: "CURRENT",
			WritableRange: "CURRENT",
		},
		{
			RoleID:        gzr.ID,
			TargetOrgID:   yxo.ID,
			ReadableRange: "CURRENT",
			WritableRange: "CURRENT",
		},
	}
	for _, item := range gzrDP {
		tx.Create(&item)
	}
	// 新增用户
	gz01 := &User{
		Username: "gz01",
		Password: MD5("12345"),
		Name:     "广州用户01",
		RoleAssigns: []RoleAssign{
			{
				RoleID: gzr.ID,
			},
		},
	}
	tx.Create(gz01)
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
		if C().GetBool("mock") {
			createMockData(tx)
		}
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
		desc := GetPrivilegesDesc(c)
		db = DB().Where("id in (?)", desc.ReadableOrgIDs).Find(&data)
	}
	if errs := db.GetErrors(); len(errs) > 0 {
		return &data, errors.New(L(c, "查询组织失败"))
	}
	return &data, nil
}

// GetUserWithRoles
var GetUserWithRoles = func(uid uint) (*User, error) {
	// 查询用户档案
	var user User
	if errs := DB().Where("id = ?", uid).Preload("RoleAssigns").First(&user).GetErrors(); len(errs) > 0 || user.ID == 0 {
		ERROR(errs)
		return &user, errors.New("查询用户失败")
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
		roleMap = make(map[uint]Role)
	)
	if errs := DB().Where("id in (?)", roleIDs).Preload("OperationPrivileges").Preload("DataPrivileges").Find(&roles).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		return &user, errors.New("查询角色失败")
	}
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	// 重新赋值
	for index, assign := range user.RoleAssigns {
		role := roleMap[assign.RoleID]
		assign.Role = &role
		user.RoleAssigns[index] = assign
	}
	return &user, nil
}

// ExecOrgLogin
func ExecOrgLogin(sign *SignContext, orgID uint) (*Org, error) {
	var orgData Org
	if errs := DB().Where("id = ?", orgID).First(&orgData).GetErrors(); len(errs) > 0 || orgData.ID == 0 {
		ERROR(errs)
		return &orgData, errors.New("组织不存在")
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
		return &orgData, errors.New("创建组织登入记录失败")
	}
	// 缓存secret至redis
	key := RedisKeyBuilder(RedisOrgKey, signOrg.Token)
	value := Stringify(&signOrg)
	if _, err := RedisClient.Set(key, value, time.Second*time.Duration(ExpiresSeconds)).Result(); err != nil {
		ERROR(err)
	}
	return &orgData, nil
}

func defaultLoginHandler(c *Context) (jwt.MapClaims, uint, error) {
	body := struct {
		Username string
		Password string
	}{}
	// 解析请求参数
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		ERROR(err)
		return nil, 0, errors.New("解析请求体失败")
	}
	// 检测账号是否存在
	var user User
	if err := DB().Where(&User{Username: body.Username}).First(&user).Error; err != nil {
		ERROR(err)
		return nil, 0, errors.New("用户不存在")
	}
	// 检测账号是否有效
	if user.Disable {
		return nil, 0, errors.New("该用户已被禁用")
	}
	// 检测密码是否正确
	body.Password = strings.ToLower(body.Password)
	if !CompareHashAndPassword(user.Password, body.Password) {
		return nil, 0, errors.New("账号密码不一致")
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
	payload = SetPayloadAttrs(payload, &user)
	return payload, user.ID, nil
}

// SetPayloadAttrs
var SetPayloadAttrs = func(payload jwt.MapClaims, user *User) jwt.MapClaims {
	return payload
}

// Sys
func Sys() *Mod {
	return &Mod{
		Code: "sys",
		Models: []interface{}{
			&User{},
			&Org{},
			&RoleAssign{},
			&Role{},
			&OperationPrivileges{},
			&DataPrivileges{},
			&Menu{},
			&Dict{},
			&DictValue{},
			&File{},
			&SignOrg{},
			&Param{},
			&Metadata{},
			&MetadataField{},
			&Route{},
			&Language{},
		},
		Middlewares: gin.HandlersChain{
			OrgMiddleware,
		},
		Routes: RoutesInfo{
			OrgLoginRoute,
			OrgListRoute,
			OrgCurrentRoute,
			UserRoleAssigns,
			UserMenusRoute,
			UploadRoute,
			AuthRoute,
			MetaRoute,
			EnumRoute,
			ModelDocsRoute,
		},
		AfterImport: initSys,
	}
}
