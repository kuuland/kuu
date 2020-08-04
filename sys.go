package kuu

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

const initCode = "sys:init"

var (
	rootUser *User
	rootOrg  *Org
	rootRole *Role
)

// RootUID
func RootUID() uint {
	return 1
}

// RootOrgID
func RootOrgID() uint {
	return 1
}

// RootRoleID
func RootRoleID() uint {
	return 1
}

// RootUser
func RootUser() *User {
	if rootUser == nil {
		db := DB().Where("id = ?", 1).First(rootUser)
		if err := db.Error; err != nil {
			ERROR(err)
		}
	}
	return rootUser
}

// RootOrg
func RootOrg() *Org {
	if rootOrg == nil {
		db := DB().Where("id = ?", 1).First(rootOrg)
		if err := db.Error; err != nil {
			ERROR(err)
		}
	}
	return rootOrg
}

// RootRole
func RootRole() *Role {
	if rootRole == nil {
		db := DB().Where("id = ?", 1).First(rootRole)
		if err := db.Error; err != nil {
			ERROR(err)
		}
	}
	return rootRole
}

func preflight() bool {
	var param Param
	DB().Where(&Param{Code: initCode, IsBuiltIn: null.NewBool(true, true)}).Find(&param)
	if param.ID != 0 {
		return true
	}
	return false
}

func initSys() {
	if !preflight() {
		// 初始化预置数据
		err := WithTransaction(func(tx *gorm.DB) error {
			// 初始化预置用户
			createRootUser(tx)
			// 初始化预置组织
			createRootOrg(tx)
			// 初始化字典、菜单
			createPresetMenus(tx)
			// 初始化预置用户权限
			createRootPrivileges(tx)
			// 保存初始化标记
			param := Param{
				Code:      initCode,
				IsBuiltIn: null.NewBool(true, true),
				Name:      "System initialization label",
				Value:     "ok",
			}
			tx.Create(&param)
			return tx.Error
		})
		if err != nil {
			PANIC("failed to initialize preset data: %s", err.Error())
		}
	}
}

func createRootUser(tx *gorm.DB) {
	root := User{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		OrgID:       RootOrgID(),
		Username:    "root",
		Name:        "Default",
		Password:    MD5("kuu"),
		IsBuiltIn:   null.NewBool(true, true),
	}
	tx.Create(&root)
	rootUser = &root
}

func createRootOrg(tx *gorm.DB) {
	root := Org{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		Code:        "default",
		Name:        "Default",
		IsBuiltIn:   null.NewBool(true, true),
	}
	tx.Create(&root)
	rootOrg = &root
}

func createRootPrivileges(tx *gorm.DB) {
	// 创建角色
	rootRole := &Role{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		OrgID:       RootOrgID(),
		Code:        "root_role",
		Name:        "Default",
		IsBuiltIn:   null.NewBool(true, true),
	}
	tx.Create(rootRole)
	// 创建数据权限记录
	tx.Create(&DataPrivileges{
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		RoleID:        rootRole.ID,
		TargetOrgID:   RootOrgID(),
		ReadableRange: DataScopeCurrentFollowing,
		WritableRange: DataScopeCurrentFollowing,
	})
	// 创建分配记录
	tx.Create(&RoleAssign{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		RoleID: rootRole.ID,
		UserID: RootUID(),
	})
}

func createPresetMenus(tx *gorm.DB) {
	rootMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Name:      "Default",
		LocaleKey: "menu_default",
		Sort:      100,
	}
	tx.Create(&rootMenu)
	sysMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       rootMenu.ID,
		Name:      "System Management",
		LocaleKey: "menu_sys_mgr",
		Icon:      "setting",
		Sort:      100,
	}
	tx.Create(&sysMenu)
	orgMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       sysMenu.ID,
		Name:      "Organization Management",
		LocaleKey: "menu_org_mgr",
		Icon:      "apartment",
		Sort:      100,
	}
	tx.Create(&orgMenu)
	userMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       orgMenu.ID,
		Name:      "User Management",
		LocaleKey: "menu_user_doc",
		Icon:      "user",
		URI:       null.NewString("/sys/user", true),
		Sort:      100,
	}
	tx.Create(&userMenu)
	sysOrgMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       orgMenu.ID,
		Name:      "Organization Management",
		LocaleKey: "menu_org_doc",
		Icon:      "cluster",
		URI:       null.NewString("/sys/org", true),
		Sort:      200,
	}
	tx.Create(&sysOrgMenu)
	permissionMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       sysMenu.ID,
		Name:      "Authorization Management",
		LocaleKey: "menu_auth_mgr",
		Icon:      "dropbox",
		Sort:      200,
	}
	tx.Create(&permissionMenu)
	roleMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       permissionMenu.ID,
		Name:      "Role Management",
		LocaleKey: "menu_role_doc",
		Icon:      "team",
		URI:       null.NewString("/sys/role", true),
		Sort:      100,
	}
	tx.Create(&roleMenu)
	settingMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       sysMenu.ID,
		Name:      "System Settings",
		LocaleKey: "menu_sys_settings",
		Icon:      "tool",
		Sort:      300,
	}
	tx.Create(&settingMenu)
	menuMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Menu Management",
		LocaleKey: "menu_menu_doc",
		Icon:      "bars",
		URI:       null.NewString("/sys/menu", true),
		Sort:      100,
	}
	tx.Create(&menuMenu)
	importMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Import Management",
		LocaleKey: "menu_import_doc",
		Icon:      "import",
		URI:       null.NewString("/sys/import", true),
		Sort:      200,
	}
	tx.Create(&importMenu)
	authorityMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Authority Management",
		LocaleKey: "menu_authority_doc",
		Icon:      "key",
		URI:       null.NewString("/sys/permission", true),
		Sort:      300,
	}
	tx.Create(&authorityMenu)
	paramMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Parameter Management",
		LocaleKey: "menu_param_doc",
		Icon:      "profile",
		URI:       null.NewString("/sys/param", true),
		Sort:      400,
	}
	tx.Create(&paramMenu)
	monitorMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "System Monitor",
		LocaleKey: "menu_sys_monitor",
		Icon:      "monitor",
		URI:       null.NewString("/sys/statics", true),
		Sort:      500,
	}
	tx.Create(&monitorMenu)
	auditMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Audit Logs",
		LocaleKey: "menu_audit_logs",
		Icon:      "book",
		URI:       null.NewString("/sys/audit", true),
		Sort:      600,
	}
	tx.Create(&auditMenu)
	fileMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "File Management",
		LocaleKey: "menu_file_doc",
		Icon:      "file",
		URI:       null.NewString("/sys/file", true),
		Sort:      700,
	}
	tx.Create(&fileMenu)
	i18nMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Internationalization",
		LocaleKey: "menu_i18n",
		Icon:      "global",
		URI:       null.NewString("/sys/i18n", true),
		Sort:      800,
	}
	tx.Create(&i18nMenu)
	messageMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Message Center",
		LocaleKey: "menu_message",
		Icon:      "message",
		URI:       null.NewString("/sys/message", true),
		Sort:      900,
	}
	tx.Create(&messageMenu)
	metadataMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Metadata Management",
		LocaleKey: "menu_metadata",
		Icon:      "appstore",
		URI:       null.NewString("/sys/metadata", true),
		Sort:      1000,
	}
	tx.Create(&metadataMenu)
}

// GetLoginableOrgs
func GetLoginableOrgs(c *Context, uid uint) ([]Org, error) {
	var (
		data []Org
	)
	sc, err := c.DecodedContext()
	if err != nil {
		return nil, err
	}
	if desc := GetPrivilegesDesc(sc); desc != nil {
		if err := DB().Where("id in (?)", desc.LoginableOrgIDs).Find(&data).Error; err != nil {
			return nil, err
		}
	}
	return data, nil
}

func GetActOrg(c *Context, actOrgID uint) (actOrg Org, err error) {
	if actOrgID != 0 && c.PrisDesc.IsReadableOrgID(actOrgID) {
		err = c.IgnoreAuth().DB().First(&actOrg, "id = ?", actOrgID).Error
	} else {
		var list []Org
		if list, err = GetLoginableOrgs(c, c.SignInfo.UID); err != nil {
			return
		}
		if len(list) > 0 {
			actOrg = list[0]
		}
	}
	return
}

// GetUserWithRoles
var GetUserWithRoles = func(uid uint) (*User, error) {
	// 查询用户档案
	var user User
	if err := DB().Where("id = ?", uid).Preload("RoleAssigns").First(&user).Error; err != nil {
		return &user, err
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
	if err := DB().Where("id in (?)", roleIDs).Preload("OperationPrivileges").Preload("DataPrivileges").Find(&roles).Error; err != nil {
		return &user, err
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

func getFailedTimesKey(username string) string {
	return strings.ToLower(fmt.Sprintf("login_%s_failed_times", username))
}

func failedTimesValid(times int) bool {
	return times > C().DefaultGetInt("captchaFailedTimes", 3)
}

func defaultLoginHandler(c *Context) (resp *LoginHandlerResponse) {
	body := struct {
		Username     string
		Password     string
		CaptchaID    string `json:"captcha_id"`
		CaptchaValue string `json:"captcha_val"`
	}{}
	// 解析请求参数
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return &LoginHandlerResponse{Error: err}
	}
	// 判断是否需要需要校验验证码
	failedTimesKey := getFailedTimesKey(body.Username)
	failedTimes := GetCacheInt(failedTimesKey)
	cacheFailedTimes := func() {
		SetCacheInt(failedTimesKey, failedTimes+1)
	}
	resp = &LoginHandlerResponse{
		Username: body.Username,
		Password: body.Password,
	}
	if failedTimesValid(failedTimes) {
		// 校验验证码
		if body.CaptchaID == "" {
			body.CaptchaID = ParseCaptchaID(c)
		}
		if body.CaptchaValue == "" {
			body.CaptchaValue = ParseCaptchaValue(c)
		}
		verifyResult := VerifyCaptcha(body.CaptchaID, body.CaptchaValue)
		if !verifyResult {
			resp.Error = fmt.Errorf("incorrect captcha code: input_id = %s, input_value = %s", body.CaptchaID, body.CaptchaValue)
			resp.LocaleMessageID = "incorrect_captcha_code"
			resp.LocaleMessageDefaultText = "Incorrect captcha code."
			return
		}
	}
	// 检测账号是否存在
	var user User
	if err := DB().Where(&User{Username: body.Username}).First(&user).Error; err != nil {
		resp.Error = err
		cacheFailedTimes()
		return
	}
	// 检测账号是否有效
	if user.Disable.Bool {
		resp.Error = fmt.Errorf("account has been disabled: %v", user.ID)
		resp.LocaleMessageID = "acc_account_disabled"
		resp.LocaleMessageDefaultText = "Account has been disabled."
		cacheFailedTimes()
		return
	}
	// 检测密码是否正确
	body.Password = strings.ToLower(body.Password)
	if err := CompareHashAndPassword(user.Password, body.Password); err != nil {
		resp.Error = err
		cacheFailedTimes()
		return
	}
	resp.Payload = jwt.MapClaims{
		"UID":       user.ID,
		"Username":  user.Username,
		"Name":      user.Name,
		"Avatar":    user.Avatar,
		"Sex":       user.Sex,
		"Mobile":    user.Mobile,
		"Email":     user.Email,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	}
	resp.Payload = SetPayloadAttrs(resp.Payload, &user)
	// 处理Lang参数
	if user.Lang == "" {
		user.Lang = c.Lang()
	}
	resp.Payload["Lang"] = user.Lang
	resp.Lang = user.Lang
	resp.UID = user.ID
	return
}

// SetPayloadAttrs
var SetPayloadAttrs = func(payload jwt.MapClaims, user *User) jwt.MapClaims {
	return payload
}

// GetUserFromCache
func GetUserFromCache(uid uint) (user User) {
	cacheKey := fmt.Sprintf("user_%d", uid)
	if info := GetCacheString(cacheKey); info != "" {
		_ = JSONParse(info, &user)
	} else {
		if err := DB().First(&user, &User{ID: uid}).Error; err == nil {
			SetCacheString(cacheKey, JSONStringify(user))
		}
	}
	return
}

// Sys
func Sys() *Mod {
	return &Mod{
		Code:       "sys",
		Middleware: HandlersChain{},
		Models: []interface{}{
			&ImportRecord{},
			&User{},
			&Org{},
			&RoleAssign{},
			&Role{},
			&OperationPrivileges{},
			&DataPrivileges{},
			&Menu{},
			&File{},
			&Param{},
			&Metadata{},
			&MetadataField{},
			&Route{},
		},
		Routes: RoutesInfo{
			OrgLoginableRoute,
			OrgSwitchRoute,
			UserRoleAssigns,
			UserMenusRoute,
			UploadRoute,
			ImportRoute,
			ImportTemplateRoute,
			AuthRoute,
			MetaRoute,
			EnumRoute,
			CaptchaRoute,
			ModelDocsRoute,
			ModelWSRoute,
			LangSwitchRoute,
			LoginAsRoute,
			LoginAsUsersRoute,
			LoginAsOutRoute,
			IntlLanguagesRoute,
			IntlMessagesRoute,
			IntlMessagesSaveRoute,
			IntlMessagesUploadRoute,
		},
		AfterImport: initSys,
	}
}
