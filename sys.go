package kuu

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
	if preflight() {
		return
	}
	IgnoreAuth()
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
		// 初始化国际化配置
		createPresetLanguageMessages(tx)
		// 保存初始化标记
		param := Param{
			Model: Model{
				CreatedByID: RootUID(),
				UpdatedByID: RootUID(),
				OrgID:       RootOrgID(),
			},
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
	IgnoreAuth(true)
}

func createRootUser(tx *gorm.DB) {
	root := User{
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		Username:  "root",
		Name:      "Default",
		Password:  MD5("kuu"),
		IsBuiltIn: null.NewBool(true, true),
	}
	tx.Create(&root)
	rootUser = &root
}

func createRootOrg(tx *gorm.DB) {
	root := Org{
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		Code: "default",
		Name: "Default",
	}
	tx.Create(&root)
	rootOrg = &root
}

func createRootPrivileges(tx *gorm.DB) {
	// 创建角色
	rootRole := &Role{
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		Code:      "root_role",
		Name:      "Default",
		IsBuiltIn: null.NewBool(true, true),
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
		Model: Model{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
			OrgID:       RootOrgID(),
		},
		RoleID: rootRole.ID,
		UserID: RootUID(),
	})
}

func createPresetLanguageMessages(tx *gorm.DB) {
	// 注册国际化语言
	tx.Create(&Language{
		LangCode: "en-US",
		LangName: "English",
	})
	tx.Create(&Language{
		LangCode: "zh-CN",
		LangName: "简体中文",
	})
	tx.Create(&Language{
		LangCode: "zh-TW",
		LangName: "繁體中文",
	})
	// 注册国际化消息条目
	register := NewLangRegister(tx)
	// 接口
	register.SetKey("acc_token_failed").Add("Token signing failed", "令牌签发失败", "令牌簽發失敗")
	register.SetKey("acc_logout_failed").Add("Logout failed", "登出失败", "登出失敗")
	register.SetKey("acc_token_expired").Add("Token has expired", "令牌已过期", "令牌已過期")
	register.SetKey("apikeys_failed").Add("Create Access Key failed", "创建访问密钥失败", "創建訪問密鑰失敗")
	register.SetKey("kuu_up").Add("{{time}}", "{{time}}", "{{time}}")
	register.SetKey("acc_please_login").Add("Please login", "请登录", "請登錄")
	register.SetKey("acc_session_expired").Add("Login session has expired", "登录会话已过期", "登錄會話已過期")
	register.SetKey("org_login_failed").Add("Organization login failed", "组织登入失败", "組織登入失敗")
	register.SetKey("org_query_failed").Add("Organization query failed", "组织列表查询失败", "組織列表查詢失敗")
	register.SetKey("org_not_found").Add("Organization not found", "组织不存在", "組織不存在")
	register.SetKey("role_assigns_failed").Add("User roles query failed", "用户角色查询失败", "用戶角色查詢失敗")
	register.SetKey("user_menus_failed").Add("User menus query failed", "用户菜单查询失败", "用戶菜單查詢失敗")
	register.SetKey("upload_failed").Add("Upload file failed", "文件上传失败", "文件上傳失敗")
	register.SetKey("auth_failed").Add("Authentication failed", "鉴权失败", "鑒權失敗")
	register.SetKey("model_docs_failed").Add("Model document query failed", "默认接口文档查询失败", "默認接口文檔查詢失敗")
	register.SetKey("lang_switch_failed").Add("Language switching failed", "语言切换失败", "語言切換失敗")
	register.SetKey("lang_msgs_failed").Add("Query messages failed", "查询国际化配置失败", "查詢國際化配置失敗")
	register.SetKey("lang_trans_query_failed").Add("Query translation list failed", "查询国际化翻译列表失败", "查詢國際化翻譯列表失敗")
	register.SetKey("lang_trans_save_failed").Add("Save locale messages failed", "保存国际化配置失败", "保存國際化配置失敗")
	register.SetKey("login_failed").Add("Login failed", "登录失败", "登錄失敗")
	register.SetKey("rest_update_failed").Add("Update failed", "更新失败", "更新失敗")
	register.SetKey("rest_query_failed").Add("Query failed", "查询失败", "查詢失敗")
	register.SetKey("rest_delete_failed").Add("Delete failed", "删除失败", "刪除失敗")
	register.SetKey("rest_create_failed").Add("Create failed", "新增失败", "新增失敗")
	// 菜单
	register.SetKey("menu_default").Add("Default", "默认菜单", "默認菜單")
	register.SetKey("menu_sys_mgr").Add("System Management", "系统管理", "系統管理")
	register.SetKey("menu_org_mgr").Add("Organization Management", "组织管理", "組織管理")
	register.SetKey("menu_user_doc").Add("User", "用户", "用戶")
	register.SetKey("menu_org_doc").Add("Organization", "组织", "組織")
	register.SetKey("menu_auth_mgr").Add("Authorization Management", "权限管理", "權限管理")
	register.SetKey("menu_role_doc").Add("Role", "角色", "角色")
	register.SetKey("menu_sys_settings").Add("System Settings", "系统设置", "系統設置")
	register.SetKey("menu_menu_doc").Add("Menu", "菜单", "菜單")
	register.SetKey("menu_param_doc").Add("Parameter", "参数", "參數")
	register.SetKey("menu_audit_logs").Add("Audit Logs", "审计", "審計")
	register.SetKey("menu_file_doc").Add("File", "文件", "文件")
	register.SetKey("menu_i18n").Add("Internationalization", "国际化", "國際化")
	register.SetKey("menu_message").Add("Message", "消息", "消息")
	// Fano
	register.SetKey("fano_table_actions_add").Add("Add", "新增", "新增")
	register.SetKey("fano_table_actions_del").Add("Del", "删除", "刪除")
	register.SetKey("fano_table_actions_cols").Add("Columns", "隐藏列", "隱藏列")
	register.SetKey("fano_table_actions_filter").Add("Filter", "过滤", "過濾")
	register.SetKey("fano_table_actions_sort").Add("Sort", "排序", "排序")
	register.SetKey("fano_table_actions_export").Add("Export", "导出", "導出")
	register.SetKey("fano_table_actions_refresh").Add("Refresh", "刷新", "刷新")
	register.SetKey("fano_table_actions_expand").Add("Expand All", "全部展开", "全部展開")
	register.SetKey("fano_table_actions_collapse").Add("Collapse All", "全部折叠", "全部折疊")
	register.SetKey("fano_table_tabs_form").Add("Form", "表单详情", "表單詳情")
	register.SetKey("fano_table_tabs_fullscreen").Add("Full Screen", "全屏", "全屏")
	register.SetKey("fano_table_tabs_close").Add("Close", "关闭", "關閉")
	register.SetKey("fano_table_total").Add("Total {{total}} items", "{{total}} 条记录", "{{total}} 條記錄")
	register.SetKey("fano_table_save_failed").Add("Save failed", "保存失败", "保存失敗")
	register.SetKey("fano_table_save_success").Add("Successfully saved", "保存成功", "保存成功")
	register.SetKey("fano_table_fill_form").Add("Please fill the form", "请填表单信息", "請填寫表單信息")
	register.SetKey("fano_table_del_selectrows").Add("Please select the rows you want to delete", "请先选择需要删除的行", "請先選擇需要刪除的行")
	register.SetKey("fano_table_del_popconfirm").Add("Are you sure to delete?", "确认删除吗？", "確認刪除嗎？")
	register.SetKey("fano_table_row_action_edit").Add("Edit this row", "编辑行", "編輯行")
	register.SetKey("fano_table_row_action_del").Add("Delete this row", "删除行", "刪除行")
	register.SetKey("fano_table_cols_actions").Add("Actions", "操作", "操作")
	register.SetKey("fano_table_filter_condtype_before").Add("Query data that meets", "筛选出符合下面", "篩選出符合下面")
	register.SetKey("fano_table_filter_condtype_all").Add("ALL", "所有", "所有")
	register.SetKey("fano_table_filter_condtype_one").Add("ONE", "任一", "任一")
	register.SetKey("fano_table_filter_condtype_after").Add("of the following rules", "条件的数据", "條件的數據")
	register.SetKey("fano_table_filter_addrule").Add("Add rule", "添加条件", "添加條件")
	register.SetKey("fano_table_filter_delrule").Add("Delete this rule", "删除条件", "刪除條件")
	register.SetKey("fano_table_filter_operators_eq").Add("Equal", "等于", "等於")
	register.SetKey("fano_table_filter_operators_ne").Add("NOT Equal", "不等于", "不等於")
	register.SetKey("fano_table_filter_operators_gt").Add("Greater Than", "大于", "大於")
	register.SetKey("fano_table_filter_operators_gte").Add("Greater Than or Equal", "大于等于", "大於等於")
	register.SetKey("fano_table_filter_operators_lt").Add("Less Than", "小于", "小於")
	register.SetKey("fano_table_filter_operators_lte").Add("Less Than or Equal", "小于等于", "小於等於")
	register.SetKey("fano_table_filter_operators_like").Add("Contains", "包含", "包含")
	register.SetKey("fano_table_filter_operators_null").Add("IS NULL", "为空", "為空")
	register.SetKey("fano_table_filter_operators_notnull").Add("IS NOT NULL", "非空", "非空")
	register.SetKey("fano_table_filter_submit").Add("Filter now", "筛选", "篩選")
	register.SetKey("fano_table_sort_addrule").Add("Add rule", "添加条件", "添加條件")
	register.SetKey("fano_table_sort_delrule").Add("Delete this rule", "删除条件", "刪除條件")
	register.SetKey("fano_table_sort_asc").Add("Ascending", "升序", "升序")
	register.SetKey("fano_table_sort_desc").Add("Descending", "降序", "降序")
	register.SetKey("fano_table_sort_submit").Add("Sort now", "排序", "排序")
	register.SetKey("fano_placeholder_choose").Add("Please choose {{name}}", "请选择{{name}}", "請選擇{{name}}")
	register.SetKey("fano_placeholder_input").Add("Please input {{name}}", "请输入{{name}}", "請輸入{{name}}")
	register.SetKey("fano_placeholder_keyword").Add("Please enter a keyword", "请输入关键字", "請輸入關鍵字")
	register.SetKey("fano_form_btnsubmit").Add("Submit", "提交", "提交")
	register.SetKey("fano_form_btncancel").Add("Cancel", "取消", "取消")
	// Kuu UI
	register.SetKey("kuu_navbar_profile").Add("Profile", "个人中心", "個人中心")
	register.SetKey("kuu_navbar_changepass").Add("Change password", "修改密码", "修改密碼")
	register.SetKey("kuu_navbar_languages").Add("Languages", "语言切换", "語言切換")
	register.SetKey("kuu_navbar_apikeys").Add("API & Keys", "API & Keys", "API & Keys")
	register.SetKey("kuu_navbar_logout").Add("Logout", "退出登录", "退出登錄")
	register.SetKey("kuu_i18n_key").Add("Key", "国际化键", "國際化鍵")
	register.SetKey("kuu_i18n_keyword_placeholder").Add("Search keywords", "输入关键字", "輸入關鍵字")
	register.SetKey("kuu_i18n_actions_languages").Add("Languages", "语言管理", "語言管理")
	register.SetKey("kuu_i18n_languages_langcode").Add("Language code", "语言管理", "語言管理")
	register.SetKey("kuu_i18n_languages_langname").Add("Language name", "语言管理", "語言管理")
	register.SetKey("lang_list_save_failed").Add("Save languages failed", "保存语言列表失败", "保存語言列表失敗")
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
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
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
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
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
		Icon:      "appstore",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
	}
	tx.Create(&orgMenu)
	userMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       orgMenu.ID,
		Name:      "User",
		LocaleKey: "menu_user_doc",
		Icon:      "user",
		URI:       "/sys/user",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&userMenu)
	sysOrgMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       orgMenu.ID,
		Name:      "Organization",
		LocaleKey: "menu_org_doc",
		Icon:      "cluster",
		URI:       "/sys/org",
		Sort:      200,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
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
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
	}
	tx.Create(&permissionMenu)
	roleMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       permissionMenu.ID,
		Name:      "Role",
		LocaleKey: "menu_role_doc",
		Icon:      "team",
		URI:       "/sys/role",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
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
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
	}
	tx.Create(&settingMenu)
	menuMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Menu",
		LocaleKey: "menu_menu_doc",
		Icon:      "bars",
		URI:       "/sys/menu",
		Sort:      100,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&menuMenu)
	paramMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Parameter",
		LocaleKey: "menu_param_doc",
		Icon:      "profile",
		URI:       "/sys/param",
		Sort:      200,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&paramMenu)
	auditMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Audit Logs",
		LocaleKey: "menu_audit_logs",
		Icon:      "book",
		URI:       "/sys/audit",
		Sort:      400,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&auditMenu)
	fileMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "File",
		LocaleKey: "menu_file_doc",
		Icon:      "file",
		URI:       "/sys/file",
		Sort:      500,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
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
		URI:       "/sys/i18n",
		Sort:      600,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&i18nMenu)
	messageMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Message",
		LocaleKey: "menu_message",
		Icon:      "message",
		URI:       "/sys/message",
		Sort:      700,
		Type:      "menu",
		IsBuiltIn: null.NewBool(true, true),
		Closeable: null.NewBool(true, true),
	}
	tx.Create(&messageMenu)
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
	return &data, db.Error
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

// ExecOrgLogin
func ExecOrgLogin(sign *SignContext, orgID uint) (*Org, error) {
	var orgData Org
	if err := DB().Where("id = ?", orgID).First(&orgData).Error; err != nil {
		return &orgData, err
	}
	// 新增登入记录
	signOrg := SignOrg{
		UID:   sign.UID,
		Token: sign.Token,
	}
	signOrg.OrgID = orgData.ID
	signOrg.CreatedByID = sign.UID
	signOrg.UpdatedByID = sign.UID
	if err := DB().Create(&signOrg).Error; err != nil {
		return &orgData, err
	}
	// 缓存secret至redis
	key := RedisKeyBuilder(RedisOrgKey, signOrg.Token)
	value := Stringify(&signOrg)
	if _, err := RedisClient.Set(key, value, time.Second*time.Duration(ExpiresSeconds)).Result(); err != nil {
		ERROR(err)
	}
	return &orgData, nil
}

func defaultLoginHandler(c *Context) (payload jwt.MapClaims, uid uint) {
	body := struct {
		Username string
		Password string
	}{}
	failedMessage := L("login_failed", "Login failed")
	// 解析请求参数
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	// 检测账号是否存在
	var user User
	if err := DB().Where(&User{Username: body.Username}).First(&user).Error; err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	// 检测账号是否有效
	if user.Disable.Bool {
		c.STDErr(failedMessage, errors.New("account has been disabled"))
		return
	}
	// 检测密码是否正确
	body.Password = strings.ToLower(body.Password)
	if !CompareHashAndPassword(user.Password, body.Password) {
		c.STDErr(failedMessage, errors.New("inconsistent account password"))
		return
	}
	payload = jwt.MapClaims{
		"UID":       user.ID,
		"Username":  user.Username,
		"Name":      user.Name,
		"Avatar":    user.Avatar,
		"Sex":       user.Sex,
		"Mobile":    user.Mobile,
		"Email":     user.Email,
		"Lang":      user.Lang,
		"IsBuiltIn": user.IsBuiltIn,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	}
	payload = SetPayloadAttrs(payload, &user)
	uid = user.ID
	if user.Lang == "" {
		user.Lang = ParseLang(c.Context)
	}
	c.SetCookie(RequestLangKey, user.Lang, ExpiresSeconds, "/", "", false, true)
	return
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
			&File{},
			&SignOrg{},
			&Param{},
			&Metadata{},
			&MetadataField{},
			&Route{},
			&Language{},
			&LanguageMessage{},
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
			LangmsgsRoute,
			LangtransGetRoute,
			LangtransPostRoute,
			LanglistPostRoute,
		},
		AfterImport: initSys,
	}
}
