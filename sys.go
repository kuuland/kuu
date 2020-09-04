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

func initSys() error {
	// 初始化预置数据
	err := WithTransaction(func(tx *gorm.DB) error {
		// 初始化预置用户
		if err := createRootUser(tx); err != nil {
			return err
		}
		// 初始化预置组织
		if err := createRootOrg(tx); err != nil {
			return err
		}
		// 初始化字典、菜单
		if err := createPresetMenus(tx); err != nil {
			return err
		}
		// 初始化预置用户权限
		if err := createRootPrivileges(tx); err != nil {
			return err
		}
		// 保存初始化标记
		param := Param{
			Code:      initCode,
			IsBuiltIn: null.NewBool(true, true),
			Name:      "System initialization flag",
			Value:     "ok",
		}
		return tx.Create(&param).Error
	})
	return err
}

func createRootUser(tx *gorm.DB) error {
	root := User{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		OrgID:       RootOrgID(),
		Username:    "root",
		Name:        "Default",
		Password:    MD5("kuu"),
		IsBuiltIn:   null.NewBool(true, true),
	}
	rootUser = &root
	return tx.Create(rootUser).Error
}

func createRootOrg(tx *gorm.DB) error {
	root := Org{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		Code:        "default",
		Name:        "Default",
		IsBuiltIn:   null.NewBool(true, true),
	}
	rootOrg = &root
	return tx.Create(rootOrg).Error
}

func createRootPrivileges(tx *gorm.DB) error {
	// 创建角色
	rootRole := &Role{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		OrgID:       RootOrgID(),
		Code:        "root_role",
		Name:        "Root Role",
		IsBuiltIn:   null.NewBool(true, true),
	}
	if err := tx.Create(rootRole).Error; err != nil {
		return err
	}
	// 创建数据权限记录
	if err := tx.Create(&DataPrivileges{
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		RoleID:        rootRole.ID,
		TargetOrgID:   RootOrgID(),
		ReadableRange: DataScopeCurrentFollowing,
		WritableRange: DataScopeCurrentFollowing,
	}).Error; err != nil {
		return err
	}
	// 创建分配记录
	if err := tx.Create(&RoleAssign{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		RoleID: rootRole.ID,
		UserID: RootUID(),
	}).Error; err != nil {
		return err
	}
	return nil
}

func createPresetMenus(tx *gorm.DB) error {
	menus := []Menu{
		{
			Name: "Default Menu",
			Code: "default",
		},
		// 权限管理 Authorization Management
		{
			ParentCode: null.StringFrom("default"),
			Name:       "Authorization Management",
			Code:       "auth",
			Icon:       null.StringFrom("key"),
		},
		{
			ParentCode: null.StringFrom("auth"),
			Name:       "Menu Management",
			Code:       "auth_menu",
			Icon:       null.StringFrom("bars"),
			URI:        null.NewString("/sys/menu", true),
		},
		{
			ParentCode: null.StringFrom("auth"),
			Name:       "Organization Management",
			Code:       "auth_org",
			Icon:       null.StringFrom("apartment"),
			URI:        null.NewString("/sys/org", true),
		},
		{
			ParentCode: null.StringFrom("auth"),
			Name:       "User Management",
			Code:       "auth_user",
			Icon:       null.StringFrom("user"),
			URI:        null.NewString("/sys/user", true),
		},
		{
			ParentCode: null.StringFrom("auth"),
			Name:       "Role Management",
			Code:       "auth_role",
			Icon:       null.StringFrom("team"),
			URI:        null.NewString("/sys/role", true),
		},
		{
			ParentCode: null.StringFrom("auth"),
			Name:       "Permission Management",
			Code:       "auth_permission",
			Icon:       null.StringFrom("key"),
			URI:        null.NewString("/sys/permission", true),
		},
		// 系统设置 System Management
		{
			ParentCode: null.StringFrom("default"),
			Name:       "System Management",
			Code:       "sys",
			Icon:       null.StringFrom("setting"),
		},
		{
			ParentCode: null.StringFrom("sys"),
			Name:       "Param Management",
			Code:       "sys_param",
			Icon:       null.StringFrom("profile"),
			URI:        null.NewString("/sys/param", true),
		},
		{
			ParentCode: null.StringFrom("sys"),
			Name:       "File Management",
			Code:       "sys_file",
			Icon:       null.StringFrom("file"),
			URI:        null.NewString("/sys/file", true),
		},
		{
			ParentCode: null.StringFrom("sys"),
			Name:       "Import Management",
			Code:       "sys_import",
			Icon:       null.StringFrom("import"),
			URI:        null.NewString("/sys/import", true),
		},
		{
			ParentCode: null.StringFrom("sys"),
			Name:       "Languages",
			Code:       "sys_languages",
			Icon:       null.StringFrom("global"),
			URI:        null.NewString("/sys/i18n", true),
		},
	}
	for i, item := range menus {
		item.LocaleKey = null.StringFrom(fmt.Sprintf("menu_%s", item.Code))
		item.Sort = null.IntFrom(int64((i + 1) * 100))
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
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
	// 检测账号是否禁止登录
	if user.DenyLogin.Bool {
		resp.Error = fmt.Errorf("account deny login: %v", user.ID)
		resp.LocaleMessageID = "acc_account_deny"
		resp.LocaleMessageDefaultText = "Account deny login."
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
			&Message{},
			&MessageReceipt{},
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
			JobRunRoute,
			MessagesLatestRoute,
			MessagesReadRoute,
		},
		OnInit: initSys,
	}
}
