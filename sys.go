package kuu

import (
	"errors"
	"fmt"
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
			// 初始化国际化配置
			createPresetLanguageMessages(tx)
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
	//// 启动日志序列化任务
	//_, _ = AddTask("@every 15m", func() {
	//
	//})
	//// 启动历史日志清除任务
	//_, _ = AddTask("@midnight", func() {
	//	day := 24 * time.Hour
	//	dest := day * 30 * 6
	//	time.Now().Add(-dest)
	//	//
	//	//time.Now().Sub()
	//	//DB().Model(&Log{}).Where("created_at < (now() - interval '30 minute')  and order_status = ?", models.OrderStatusPend).
	//	//	Updates(&models.Order{
	//	//		OrderStatus: models.OrderStatusCancel,
	//	//	})
	//})
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
	register.SetKey("kuu_up").Add("{{time}}", "{{time}}", "{{time}}")
	// 接口
	register.SetKey("captcha_failed").Add("Incorrect captcha code", "验证码不正确", "驗證碼不正確")
	register.SetKey("acc_token_failed").Add("Token signing failed", "令牌签发失败", "令牌簽發失敗")
	register.SetKey("acc_login_failed").Add("Login failed", "登录失败", "登錄失敗")
	register.SetKey("acc_login_disabled").Add("Account has been disabled", "账号已停用", "賬號已停用")
	register.SetKey("acc_logout_failed").Add("Logout failed", "登出失败", "登出失敗")
	register.SetKey("acc_password_failed").Add("The password you entered isn't right.", "账号密码不一致", "賬號密碼不一致")
	register.SetKey("acc_token_expired").Add("Token has expired", "令牌已过期", "令牌已過期")
	register.SetKey("sys_meta_failed").Add("Metadata does not exist: {{name}}", "元数据不存在：{{name}}", "元數據不存在：{{name}}")

	register.SetKey("apikeys_failed").Add("Create Access Key failed", "创建访问密钥失败", "創建訪問密鑰失敗")
	register.SetKey("acc_please_login").Add("Please login", "请重新登录", "請重新登錄")
	register.SetKey("acc_session_expired").Add("Login session has expired", "登录会话已过期", "登錄會話已過期")
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
	register.SetKey("lang_list_save_failed").Add("Save languages failed", "保存语言列表失败", "保存語言列表失敗")
	// Model RESTful
	register.SetKey("rest_update_failed").Add("Update failed", "更新失败", "更新失敗")
	register.SetKey("rest_query_failed").Add("Query failed", "查询失败", "查詢失敗")
	register.SetKey("rest_delete_failed").Add("Delete failed", "删除失败", "刪除失敗")
	register.SetKey("rest_create_failed").Add("Create failed", "新增失败", "新增失敗")
	register.SetKey("rest_import_failed").Add("Import failed", "导入失败", "導入失敗")
	register.SetKey("rest_export_failed").Add("Export failed", "导出失败", "導出失敗")
	// 菜单
	register.SetKey("menu_default").Add("Default", "默认菜单", "默認菜單")
	register.SetKey("menu_sys_mgr").Add("System Management", "系统管理", "系統管理")
	register.SetKey("menu_org_mgr").Add("Organization Management", "组织管理", "組織管理")
	register.SetKey("menu_user_doc").Add("User Management", "用户管理", "用戶管理")
	register.SetKey("menu_org_doc").Add("Organization Management", "组织管理", "組織管理")
	register.SetKey("menu_auth_mgr").Add("Authorization Management", "权限管理", "權限管理")
	register.SetKey("menu_role_doc").Add("Role Management", "角色管理", "角色管理")
	register.SetKey("menu_sys_settings").Add("System Settings", "系统设置", "系統設置")
	register.SetKey("menu_menu_doc").Add("Menu Management", "菜单管理", "菜單管理")
	register.SetKey("menu_param_doc").Add("Parameter Management", "参数管理", "參數管理")
	register.SetKey("menu_audit_logs").Add("Audit Logs", "系统审计", "系統審計")
	register.SetKey("menu_sys_monitor").Add("System Monitor", "系统监控", "系統監控")
	register.SetKey("menu_file_doc").Add("File Management", "文件管理", "文件管理")
	register.SetKey("menu_i18n").Add("Internationalization", "国际化", "國際化")
	register.SetKey("menu_message").Add("Message Center", "消息中心", "消息中心")
	register.SetKey("menu_metadata").Add("Metadata", "元数据", "元數據")
	// Fano
	register.SetKey("fano_table_actions_add").Add("Add", "新增", "新增")
	register.SetKey("fano_table_actions_del").Add("Del", "删除", "刪除")
	register.SetKey("fano_table_actions_cols").Add("Columns", "隐藏列", "隱藏列")
	register.SetKey("fano_table_actions_filter").Add("Filter", "过滤", "過濾")
	register.SetKey("fano_table_actions_sort").Add("Sort", "排序", "排序")
	register.SetKey("fano_table_actions_import").Add("Import", "导入", "導入")
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
	register.SetKey("fano_table_row_action_edit_text").Add("EDIT", "编辑", "編輯")
	register.SetKey("fano_table_row_action_del").Add("Delete this row", "删除行", "刪除行")
	register.SetKey("fano_table_row_action_del_text").Add("DELETE", "删除", "刪除")
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
	// Kuu Navbar
	register.SetKey("kuu_navbar_profile").Add("Profile", "个人中心", "個人中心")
	register.SetKey("kuu_navbar_changepass").Add("Change password", "修改密码", "修改密碼")
	register.SetKey("kuu_navbar_languages").Add("Languages", "语言切换", "語言切換")
	register.SetKey("kuu_navbar_apikeys").Add("API & Keys", "API & Keys", "API & Keys")
	register.SetKey("kuu_navbar_logout").Add("Logout", "退出登录", "退出登錄")
	register.SetKey("kuu_navbar_apiendpoint").Add("API Endpoint", "API服务器", "API服務器")
	register.SetKey("kuu_navbar_apiendpoint_placeholder").Add("Optional: e.g. https://kuu.example.com/api", "例如：https://kuu.example.com/api", "例如：https://kuu.example.com/api")
	// Kuu i18n
	register.SetKey("kuu_i18n_key").Add("Key", "国际化键", "國際化鍵")
	register.SetKey("kuu_i18n_keyword_placeholder").Add("Search keywords", "输入关键字", "輸入關鍵字")
	register.SetKey("kuu_i18n_key_required").Add("Key is required", "消息键是必须的", "消息鍵是必須的")
	register.SetKey("kuu_i18n_actions_languages").Add("Languages", "语言管理", "語言管理")
	register.SetKey("kuu_i18n_languages_langcode").Add("Language code", "语言编码", "語言編碼")
	register.SetKey("kuu_i18n_languages_langname").Add("Language name", "语言名称", "語言名稱")
	// Kuu Login
	register.SetKey("kuu_login_username_required").Add("Please enter your username", "请输入登录账号", "請輸入登錄賬號")
	register.SetKey("kuu_login_username_placeholder").Add("Username", "账号", "賬號")
	register.SetKey("kuu_login_password_required").Add("Please enter your password", "请输入登录密码", "請輸入登錄密碼")
	register.SetKey("kuu_login_password_placeholder").Add("Password", "密码", "密碼")
	register.SetKey("kuu_login_captcha_required").Add("Please enter the captcha", "请输入验证码", "請輸入驗證碼")
	register.SetKey("kuu_login_captcha_placeholder").Add("Captcha", "验证码", "驗證碼")
	register.SetKey("kuu_login_password_forgot").Add("Forgot your password?", "忘记密码？", "忘記密碼？")
	register.SetKey("kuu_login_remember").Add("Remember", "记住我", "記住我")
	register.SetKey("kuu_login_btn_submit").Add("Login", "登录", "登錄")
	// Kuu Layout Tabs
	register.SetKey("kuu_layout_tabs_refresh").Add("Refresh", "刷新", "刷新")
	register.SetKey("kuu_layout_tabs_close_others").Add("Close Others", "关闭其他", "關閉其他")
	register.SetKey("kuu_layout_tabs_close_left").Add("Close All to the Left", "关闭左侧", "關閉左側")
	register.SetKey("kuu_layout_tabs_close_right").Add("Close All to the Right", "关闭右侧", "關閉右側")
	// Kuu Exception
	register.SetKey("kuu_exception_403").Add("Sorry, you don't have access to this page.", "抱歉，您无权访问此页面。", "抱歉，您無權訪問此頁面。")
	register.SetKey("kuu_exception_404").Add("Sorry, the page you visited does not exist.", "抱歉，您访问的页面不存在。", "抱歉，您訪問的頁面不存在。")
	register.SetKey("kuu_exception_500").Add("Sorry, the server is reporting an error.", "抱歉，服务器报告错误。", "抱歉，服務器報告錯誤。")
	register.SetKey("kuu_exception_btn_back").Add("back to home", "回到主页", "回到主頁")
	// Kuu Org
	register.SetKey("kuu_org_unorganized").Add("You have not assigned any organization", "您尚未分配任何组织", "您尚未分配任何組織")
	register.SetKey("kuu_org_select_login").Add("Please select a login organization", "请选择登录组织", "請選擇登入組織")
	register.SetKey("kuu_org_btn_login").Add("Login", "登录", "登錄")
	register.SetKey("kuu_org_btn_login").Add("Login", "登录", "登錄")
	register.SetKey("kuu_org_name").Add("Name", "组织名称", "組織名稱")
	register.SetKey("kuu_org_code").Add("Code", "组织编码", "組織編碼")
	register.SetKey("kuu_org_sort").Add("Sort", "排序", "排序")
	register.SetKey("kuu_org_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_org_parent").Add("Parent", "上级组织", "上級組織")
	// Kuu API & Keys
	register.SetKey("kuu_apikeys_desc").Add("Description", "描述", "描述")
	register.SetKey("kuu_apikeys_desc_render").Add("User login", "用户登录", "用戶登錄")
	register.SetKey("kuu_apikeys_desc_placeholder").Add("Optional: e.g. This key is used by the cron service to trigger jobs", "任务调度中心使用此密钥来触发定时任务", "任務調度中心使用此密鑰來觸發定時任務")
	register.SetKey("kuu_apikeys_desc_required").Add("Please enter a description", "请输入描述", "請輸入描述")
	register.SetKey("kuu_apikeys_state").Add("State", "有效状态", "有效狀態")
	register.SetKey("kuu_apikeys_exp").Add("Exp", "过期时间", "過期時間")
	register.SetKey("kuu_apikeys_exp_never_exp").Add("Never Expire", "永不过期", "永不過期")
	register.SetKey("kuu_apikeys_exp_options_never").Add("Never", "永不过期", "永不過期")
	register.SetKey("kuu_apikeys_exp_options_day").Add("A day from now", "从现在开始，有效期1天", "從現在開始，有效期1天")
	register.SetKey("kuu_apikeys_exp_options_week").Add("A week from now", "从现在开始，有效期1周", "從現在開始，有效期1周")
	register.SetKey("kuu_apikeys_exp_options_month").Add("A month from now", "从现在开始，有效期1个月", "從現在開始，有效期1個月")
	register.SetKey("kuu_apikeys_exp_options_year").Add("A year from now", "从现在开始，有效期1年", "從現在開始，有效期1年")
	register.SetKey("kuu_apikeys_exp_required").Add("Please select automatic expiration time", "请选择自动过期时间", "請選擇自動過期時間")
	register.SetKey("kuu_apikeys_token_copy_copied").Add("The token has been copied", "令牌已复制", "令牌已複製")
	register.SetKey("kuu_apikeys_token_copy_tooltip").Add("Click to copy token", "点击复制令牌", "點擊複製令牌")
	register.SetKey("kuu_apikeys_token_exp_confirm").Add("Are you sure to expire this token?", "确定作废此令牌吗？", "確定作廢此令牌嗎？")
	register.SetKey("kuu_apikeys_token_exp_tooltip").Add("Expired now", "立即过期", "立即過期")
	register.SetKey("kuu_apikeys_token_enable_confirm").Add("Are you sure to re-enable the token?", "确定重新激活该令牌吗？", "確定重新激活該令牌嗎？")
	register.SetKey("kuu_apikeys_token_enable_tooltip").Add("Enable now", "立即启用", "立即啟用")
	register.SetKey("kuu_apikeys_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_apikeys_token_copy_txt").Add("Copy", "复制令牌", "複製令牌")
	register.SetKey("kuu_apikeys_token_exp_txt").Add("Expire", "立即过期", "立即過期")
	register.SetKey("kuu_apikeys_token_enable_txt").Add("Enable", "重新启用", "從新啟用")
	// Kuu Menu
	register.SetKey("kuu_menu_name").Add("Name", "菜单名称", "菜單名稱")
	register.SetKey("kuu_menu_name_required").Add("Please enter a menu name", "请输入菜单名称", "請輸入菜單名稱")
	register.SetKey("kuu_menu_uri").Add("URI", "菜单地址", "菜單地址")
	register.SetKey("kuu_menu_sort").Add("Sort", "菜单排序", "菜單排序")
	register.SetKey("kuu_menu_disable").Add("Disable", "是否禁用", "是否禁用")
	register.SetKey("kuu_menu_detail").Add("Detail", "详情", "詳情")
	register.SetKey("kuu_menu_external").Add("External link", "外部链接", "外部鏈接")
	register.SetKey("kuu_menu_defaultopen").Add("Open by default", "默认打开", "默認打開")
	register.SetKey("kuu_menu_closeable").Add("Closeable", "可关闭", "可關閉")
	register.SetKey("kuu_menu_code").Add("Permission Code", "权限编码", "權限編碼")
	register.SetKey("kuu_menu_parent").Add("Parent Menu", "父级菜单", "父級菜單")
	register.SetKey("kuu_menu_localekey").Add("Locale Key", "国际化键", "國際化鍵")
	register.SetKey("kuu_menu_localekey").Add("Locale Key", "国际化键", "國際化鍵")
	register.SetKey("kuu_menu_icon").Add("Icon", "图标", "圖標")
	register.SetKey("kuu_menu_add_submenu").Add("Add Submenu", "添加子级", "添加子級")
	// Kuu Param
	register.SetKey("kuu_param_name").Add("Name", "参数名称", "參數名稱")
	register.SetKey("kuu_param_code").Add("Code", "参数编码", "參數編碼")
	register.SetKey("kuu_param_value").Add("Value", "参数值", "參數值")
	register.SetKey("kuu_param_builtin").Add("Built-in", "是否内置", "是否內置")
	register.SetKey("kuu_param_sort").Add("Sort", "排序", "排序")
	register.SetKey("kuu_param_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_param_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_param_type").Add("Type", "值类型", "值類型")
	register.SetKey("kuu_param_type_input").Add("Input", "输入框", "輸入框")
	register.SetKey("kuu_param_type_password").Add("Password", "密码输入框", "密碼輸入框")
	register.SetKey("kuu_param_type_number").Add("Number", "数字输入框", "數字輸入框")
	register.SetKey("kuu_param_type_textarea").Add("Textarea", "多行文本", "多行文本")
	register.SetKey("kuu_param_type_editor").Add("Editor", "编辑器", "編輯器")
	register.SetKey("kuu_param_type_json").Add("JSON", "JSON编辑器", "JSON編輯器")
	register.SetKey("kuu_param_type_switch").Add("Switch", "开关", "開關")
	register.SetKey("kuu_param_type_datepicker").Add("DatePicker", "日期选择框", "日期選擇框")
	register.SetKey("kuu_param_type_rangepicker").Add("RangePicker", "范围选择框", "範圍選擇框")
	register.SetKey("kuu_param_type_monthpicker").Add("MonthPicker", "月份选择框", "月份選擇框")
	register.SetKey("kuu_param_type_weekpicker").Add("WeekPicker", "周选择框", "周選擇框")
	register.SetKey("kuu_param_type_timepicker").Add("TimePicker", "时间选择框", "時間選擇框")
	register.SetKey("kuu_param_type_upload").Add("Upload", "上传组件", "上傳組件")
	register.SetKey("kuu_param_type_color").Add("Color", "顏色组件", "顏色組件")
	register.SetKey("kuu_param_type_icon").Add("Icon", "图标组件", "圖標組件")
	// Kuu Role
	register.SetKey("kuu_role_code").Add("Code", "角色编码", "角色編碼")
	register.SetKey("kuu_role_name").Add("Name", "角色名称", "角色名稱")
	register.SetKey("kuu_role_builtin").Add("Built-in", "是否内置", "是否內置")
	register.SetKey("kuu_role_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_role_name_required").Add("Please enter a role name", "请输入角色名称", "請輸入角色名稱")
	register.SetKey("kuu_role_code_required").Add("Please enter a role code", "请输入角色编码", "請輸入角色編碼")
	register.SetKey("kuu_role_op").Add("Menu Privileges", "菜单权限", "菜單權限")
	register.SetKey("kuu_role_dp").Add("Data Privileges", "数据权限", "數據權限")
	register.SetKey("kuu_role_select_org_placeholder").Add("Please select an organization", "请选择组织", "請選擇組織")
	register.SetKey("kuu_role_readable_range_placeholder").Add("Please select a readable range", "请选择可读数据范围", "請選擇可讀數據範圍")
	register.SetKey("kuu_role_writable_range_placeholder").Add("Please select a writable range", "请选择可写数据范围", "請選擇可寫數據範圍")
	register.SetKey("kuu_role_data_range_personal").Add("PERSONAL", "个人范围", "個人範圍")
	register.SetKey("kuu_role_data_range_current").Add("CURRENT", "当前组织", "當前組織")
	register.SetKey("kuu_role_data_range_current_following").Add("CURRENT_FOLLOWING", "当前组织及以下", "當前組織及以下")
	register.SetKey("kuu_role_addrule").Add("Add rule", "添加规则", "添加規則")
	// Kuu User
	register.SetKey("kuu_user_username").Add("Username", "用户名", "用戶名")
	register.SetKey("kuu_user_name").Add("Real name", "真实姓名", "真實姓名")
	register.SetKey("kuu_user_disable").Add("Disable", "是否禁用", "是否禁用")
	register.SetKey("kuu_user_builtin").Add("Built-in", "是否内置", "是否內置")
	register.SetKey("kuu_user_createdat").Add("Created At", "创建时间", "創建時間")
	register.SetKey("kuu_user_password").Add("Password", "密码", "密碼")
	register.SetKey("kuu_user_role_assigns").Add("Role Assignments", "角色分配", "角色分配")
	register.SetKey("kuu_user_titles_notassigned").Add("Not Assigned", "未分配角色", "未分配角色")
	register.SetKey("kuu_user_titles_assigned").Add("Assigned", "已分配角色", "已分配角色")
	// Kuu Metadata
	register.SetKey("kuu_meta_code").Add("Module", "所属模块", "所屬模塊")
	register.SetKey("kuu_meta_name").Add("Name", "名称", "名稱")
	register.SetKey("kuu_meta_displayname").Add("Display Name", "显示名", "顯示名")
	register.SetKey("kuu_meta_fields_code").Add("Field Code", "字段编码", "字段編碼")
	register.SetKey("kuu_meta_fields_name").Add("Field Name", "字段名称", "字段名稱")
	register.SetKey("kuu_meta_fields_kind").Add("Kind", "字段类型", "字段類型")
	register.SetKey("kuu_meta_fields_enum").Add("Enum", "关联枚举", "關聯枚舉")
	register.SetKey("kuu_meta_fields_isref").Add("Is Ref", "是否引用", "是否引用")
	register.SetKey("kuu_meta_fields_ispassword").Add("Is Password", "是否密码", "是否密碼")
	register.SetKey("kuu_meta_fields_isarray").Add("Is Array", "是否数组", "是否數組")
	// Kuu File
	register.SetKey("kuu_file_uid").Add("UID", "文件唯一ID", "文件唯一ID")
	register.SetKey("kuu_file_class").Add("Class", "文件分类", "文件分類")
	register.SetKey("kuu_file_type").Add("Type", "文件类型", "文件類型")
	register.SetKey("kuu_file_size").Add("Size", "文件大小", "文件大小")
	register.SetKey("kuu_file_name").Add("Name", "文件名称", "文件名稱")
	register.SetKey("kuu_file_url").Add("URL", "文件地址", "文件地址")
	register.SetKey("kuu_file_createdat").Add("Created At", "上传时间", "上傳時間")
	register.SetKey("kuu_file_actions_upload").Add("Upload", "上传文件", "上傳文件")
	register.SetKey("kuu_common_org").Add("Organization", "所属组织", "所屬組織")
	_ = register.Exec(true)
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
		Type:      "menu",
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
		URI:       null.NewString("/sys/user", true),
		Sort:      100,
		Type:      "menu",
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
		URI:       null.NewString("/sys/org", true),
		Sort:      200,
		Type:      "menu",
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
		URI:       null.NewString("/sys/role", true),
		Sort:      100,
		Type:      "menu",
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
		URI:       null.NewString("/sys/menu", true),
		Sort:      100,
		Type:      "menu",
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
		URI:       null.NewString("/sys/param", true),
		Sort:      200,
		Type:      "menu",
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
		Sort:      300,
		Type:      "menu",
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
		Sort:      400,
		Type:      "menu",
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
		URI:       null.NewString("/sys/file", true),
		Sort:      500,
		Type:      "menu",
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
		Sort:      600,
		Type:      "menu",
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
		URI:       null.NewString("/sys/message", true),
		Sort:      700,
		Type:      "menu",
	}
	tx.Create(&messageMenu)
	metadataMenu := Menu{
		ModelExOrg: ModelExOrg{
			CreatedByID: RootUID(),
			UpdatedByID: RootUID(),
		},
		Pid:       settingMenu.ID,
		Name:      "Metadata",
		LocaleKey: "menu_metadata",
		Icon:      "appstore",
		URI:       null.NewString("/sys/metadata", true),
		Sort:      700,
		Type:      "menu",
	}
	tx.Create(&metadataMenu)
}

// GetSignContext
func GetSignContext(c *gin.Context) (sign *SignContext) {
	if c != nil {
		// 解析登录信息
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
		} else {
			if v, err := DecodedContext(c); err == nil {
				sign = v
			}
		}
	}
	return
}

// GetLoginableOrgs
func GetLoginableOrgs(c *gin.Context, uid uint) ([]Org, error) {
	var (
		data []Org
		db   *gorm.DB
	)
	if desc := GetPrivilegesDesc(c); desc != nil {
		db = DB().Where("id in (?)", desc.LoginableOrgIDs).Find(&data)
	}
	return data, db.Error
}

func GetActOrg(c *Context, actOrgID uint) (actOrg Org, err error) {
	if actOrgID != 0 && c.PrisDesc.IsReadableOrgID(actOrgID) {
		err = c.IgnoreAuth().DB().First(&actOrg, "id = ?", actOrgID).Error
	} else {
		var list []Org
		if list, err = GetLoginableOrgs(c.Context, c.SignInfo.UID); err != nil {
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
		resp.Error = err
		return
	}
	// 判断是否需要需要校验验证码
	failedTimesKey := getFailedTimesKey(body.Username)
	failedTimes := GetCacheInt(failedTimesKey)
	cacheFailedTimes := func() {
		SetCacheInt(failedTimesKey, failedTimes+1)
	}
	resp = &LoginHandlerResponse{
		Username:        body.Username,
		Password:        body.Password,
		LanguageMessage: c.L("acc_login_failed", "Login failed"),
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
			resp.Error = errors.New("verify captcha failed")
			resp.LanguageMessage = c.L("captcha_failed", "Incorrect captcha code")
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
		resp.LanguageMessage = c.L("acc_login_disabled", "Account has been disabled")
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
		user.Lang = ParseLang(c.Context)
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
		Parse(info, &user)
	} else {
		if err := DB().First(uid, &User{ID: uid}); err != nil {
			SetCacheString(cacheKey, Stringify(user))
		}
	}
	return
}

// Sys
func Sys() *Mod {
	return &Mod{
		Code:        "sys",
		Middlewares: gin.HandlersChain{
			//LogMiddleware,
		},
		Models: []interface{}{
			&ExcelTemplate{},
			&ExcelTemplateHeader{},
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
			&Language{},
			&LanguageMessage{},
			//&Log{},
		},
		Routes: RoutesInfo{
			OrgLoginableRoute,
			UserRoleAssigns,
			UserMenusRoute,
			UploadRoute,
			AuthRoute,
			MetaRoute,
			EnumRoute,
			CaptchaRoute,
			ModelDocsRoute,
			ModelWSRoute,
			LangmsgsRoute,
			LangtransGetRoute,
			LangtransPostRoute,
			LanglistPostRoute,
			LangtransImportRoute,
		},
		AfterImport: initSys,
	}
}
