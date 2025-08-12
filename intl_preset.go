package kuu

import (
	"github.com/kuuland/kuu/intl"
	"sync"
)

var presetIntlMessagesMu sync.RWMutex
var presetIntlMessages = map[string]map[string]string{
	"acc_incorrect_token": {
		"en":      "Incorrect token type",
		"zh-Hans": "令牌类型不正确",
		"zh-Hant": "令牌類型不正確",
	},
	"acc_account_deny": {
		"en":      "Account deny login.",
		"zh-Hans": "账号不允许登录。",
		"zh-Hant": "賬號不允許登錄。",
	},
	"acc_account_disabled": {
		"en":      "Account has been disabled.",
		"zh-Hans": "账号已被停用。",
		"zh-Hant": "賬號已被停用。",
	},
	"acc_login_failed": {
		"en":      "Login failed",
		"zh-Hans": "登录失败",
		"zh-Hant": "登錄失敗",
	},
	"acc_logout_failed": {
		"en":      "Logout failed",
		"zh-Hans": "登出失败",
		"zh-Hant": "登出失敗",
	},
	"api_permission_denied": {
		"en":      "API access denied",
		"zh-Hans": "API访问被拒绝",
		"zh-Hant": "API訪問被拒絕",
	},
	"acc_password_failed": {
		"en":      "Incorrect username or password.",
		"zh-Hans": "账号密码不一致",
		"zh-Hant": "賬號密碼不一致",
	},
	"acc_please_login": {
		"en":      "Please login",
		"zh-Hans": "请重新登录",
		"zh-Hant": "請重新登錄",
	},
	"acc_session_expired": {
		"en":      "Login session has expired",
		"zh-Hans": "登录会话已过期",
		"zh-Hant": "登錄會話已過期",
	},
	"acc_invalid_token": {
		"en":      "Invalid token.",
		"zh-Hans": "令牌不正确或已过期。",
		"zh-Hant": "令牌不正確或已過期。",
	},
	"acc_token_failed": {
		"en":      "Token signing failed",
		"zh-Hans": "令牌签发失败",
		"zh-Hant": "令牌簽發失敗",
	},
	"apikeys_failed": {
		"en":      "Create Access Key failed",
		"zh-Hans": "创建访问密钥失败",
		"zh-Hant": "創建訪問密鑰失敗",
	},
	"auth_failed": {
		"en":      "Authentication failed",
		"zh-Hans": "鉴权失败",
		"zh-Hant": "鑒權失敗",
	},
	"incorrect_captcha_code": {
		"en":      "Incorrect captcha code.",
		"zh-Hans": "验证码不正确。",
		"zh-Hant": "驗證碼不正確。",
	},
	"fano_form_btncancel": {
		"en":      "Cancel",
		"zh-Hans": "取消",
		"zh-Hant": "取消",
	},
	"fano_form_btnsubmit": {
		"en":      "Submit",
		"zh-Hans": "提交",
		"zh-Hant": "提交",
	},
	"fano_placeholder_choose": {
		"en":      "Please choose {{name}}",
		"zh-Hans": "请选择{{name}}",
		"zh-Hant": "請選擇{{name}}",
	},
	"fano_placeholder_input": {
		"en":      "Please input {{name}}",
		"zh-Hans": "请输入{{name}}",
		"zh-Hant": "請輸入{{name}}",
	},
	"fano_placeholder_keyword": {
		"en":      "Please enter a keyword",
		"zh-Hans": "请输入关键字",
		"zh-Hant": "請輸入關鍵字",
	},
	"fano_table_actions_add": {
		"en":      "Add",
		"zh-Hans": "新增",
		"zh-Hant": "新增",
	},
	"fano_table_actions_collapse": {
		"en":      "Collapse All",
		"zh-Hans": "全部折叠",
		"zh-Hant": "全部折疊",
	},
	"fano_table_actions_cols": {
		"en":      "Columns",
		"zh-Hans": "隐藏列",
		"zh-Hant": "隱藏列",
	},
	"fano_table_actions_del": {
		"en":      "Del",
		"zh-Hans": "删除",
		"zh-Hant": "刪除",
	},
	"fano_table_actions_expand": {
		"en":      "Expand All",
		"zh-Hans": "全部展开",
		"zh-Hant": "全部展開",
	},
	"fano_table_actions_export": {
		"en":      "Export",
		"zh-Hans": "导出",
		"zh-Hant": "導出",
	},
	"fano_table_actions_filter": {
		"en":      "Filter",
		"zh-Hans": "过滤",
		"zh-Hant": "過濾",
	},
	"fano_table_actions_import": {
		"en":      "Import",
		"zh-Hans": "导入",
		"zh-Hant": "導入",
	},
	"fano_table_actions_refresh": {
		"en":      "Refresh",
		"zh-Hans": "刷新",
		"zh-Hant": "刷新",
	},
	"fano_table_actions_sort": {
		"en":      "Sort",
		"zh-Hans": "排序",
		"zh-Hant": "排序",
	},
	"fano_table_cols_actions": {
		"en":      "Actions",
		"zh-Hans": "操作",
		"zh-Hant": "操作",
	},
	"fano_table_del_popconfirm": {
		"en":      "Are you sure to delete?",
		"zh-Hans": "确认删除吗？",
		"zh-Hant": "確認刪除嗎？",
	},
	"fano_table_del_selectrows": {
		"en":      "Please select the rows you want to delete",
		"zh-Hans": "请先选择需要删除的行",
		"zh-Hant": "請先選擇需要刪除的行",
	},
	"fano_table_deselect_all": {
		"en":      "Deselect All",
		"zh-Hans": "反选",
		"zh-Hant": "反選",
	},
	"fano_table_fill_form": {
		"en":      "Please fill the form",
		"zh-Hans": "请填表单信息",
		"zh-Hant": "請填寫表單信息",
	},
	"fano_table_filter_addrule": {
		"en":      "Add rule",
		"zh-Hans": "添加条件",
		"zh-Hant": "添加條件",
	},
	"fano_table_filter_condtype_after": {
		"en":      "of the following rules",
		"zh-Hans": "条件的数据",
		"zh-Hant": "條件的數據",
	},
	"fano_table_filter_condtype_all": {
		"en":      "ALL",
		"zh-Hans": "所有",
		"zh-Hant": "所有",
	},
	"fano_table_filter_condtype_before": {
		"en":      "Query data that meets",
		"zh-Hans": "筛选出符合下面",
		"zh-Hant": "篩選出符合下面",
	},
	"fano_table_filter_condtype_one": {
		"en":      "ONE",
		"zh-Hans": "任一",
		"zh-Hant": "任一",
	},
	"fano_table_filter_delrule": {
		"en":      "Delete this rule",
		"zh-Hans": "删除条件",
		"zh-Hant": "刪除條件",
	},
	"fano_table_filter_operators_eq": {
		"en":      "Equal",
		"zh-Hans": "等于",
		"zh-Hant": "等於",
	},
	"fano_table_filter_operators_gt": {
		"en":      "Greater Than",
		"zh-Hans": "大于",
		"zh-Hant": "大於",
	},
	"fano_table_filter_operators_gte": {
		"en":      "Greater Than or Equal",
		"zh-Hans": "大于等于",
		"zh-Hant": "大於等於",
	},
	"fano_table_filter_operators_like": {
		"en":      "Contains",
		"zh-Hans": "包含",
		"zh-Hant": "包含",
	},
	"fano_table_filter_operators_lt": {
		"en":      "Less Than",
		"zh-Hans": "小于",
		"zh-Hant": "小於",
	},
	"fano_table_filter_operators_lte": {
		"en":      "Less Than or Equal",
		"zh-Hans": "小于等于",
		"zh-Hant": "小於等於",
	},
	"fano_table_filter_operators_ne": {
		"en":      "NOT Equal",
		"zh-Hans": "不等于",
		"zh-Hant": "不等於",
	},
	"fano_table_filter_operators_notnull": {
		"en":      "IS NOT NULL",
		"zh-Hans": "非空",
		"zh-Hant": "非空",
	},
	"fano_table_filter_operators_null": {
		"en":      "IS NULL",
		"zh-Hans": "为空",
		"zh-Hant": "為空",
	},
	"fano_table_filter_submit": {
		"en":      "Filter now",
		"zh-Hans": "筛选",
		"zh-Hant": "篩選",
	},
	"fano_table_less": {
		"en":      "Less",
		"zh-Hans": "收起",
		"zh-Hant": "收起",
	},
	"fano_table_more": {
		"en":      "More",
		"zh-Hans": "更多",
		"zh-Hant": "更多",
	},
	"fano_table_reset": {
		"en":      "Reset",
		"zh-Hans": "重置",
		"zh-Hant": "重置",
	},
	"fano_table_row_action_del": {
		"en":      "Delete this row",
		"zh-Hans": "删除行",
		"zh-Hant": "刪除行",
	},
	"fano_table_row_action_del_text": {
		"en":      "Delete",
		"zh-Hans": "删除",
		"zh-Hant": "刪除",
	},
	"fano_table_row_action_edit": {
		"en":      "Edit this row",
		"zh-Hans": "编辑行",
		"zh-Hant": "編輯行",
	},
	"fano_table_row_action_edit_text": {
		"en":      "Edit",
		"zh-Hans": "编辑",
		"zh-Hant": "編輯",
	},
	"fano_table_save_failed": {
		"en":      "Save failed",
		"zh-Hans": "保存失败",
		"zh-Hant": "保存失敗",
	},
	"fano_table_save_success": {
		"en":      "Successfully saved",
		"zh-Hans": "保存成功",
		"zh-Hant": "保存成功",
	},
	"kuu_save_success": {
		"en":      "Saved successfully",
		"zh-Hans": "保存成功",
		"zh-Hant": "保存成功",
	},
	"kuu_copy_success": {
		"en":      "Copy successfully.",
		"zh-Hans": "复制成功",
		"zh-Hant": "複製成功",
	},
	"fano_table_search": {
		"en":      "Search",
		"zh-Hans": "搜索",
		"zh-Hant": "搜索",
	},
	"fano_table_select_all": {
		"en":      "Select All",
		"zh-Hans": "全选",
		"zh-Hant": "全選",
	},
	"fano_table_sort_addrule": {
		"en":      "Add rule",
		"zh-Hans": "添加条件",
		"zh-Hant": "添加條件",
	},
	"fano_table_sort_asc": {
		"en":      "Ascending",
		"zh-Hans": "升序",
		"zh-Hant": "升序",
	},
	"fano_table_sort_delrule": {
		"en":      "Delete this rule",
		"zh-Hans": "删除条件",
		"zh-Hant": "刪除條件",
	},
	"fano_table_sort_desc": {
		"en":      "Descending",
		"zh-Hans": "降序",
		"zh-Hant": "降序",
	},
	"fano_table_sort_submit": {
		"en":      "Sort now",
		"zh-Hans": "排序",
		"zh-Hant": "排序",
	},
	"fano_table_tabs_close": {
		"en":      "Close",
		"zh-Hans": "关闭",
		"zh-Hant": "關閉",
	},
	"fano_table_tabs_form": {
		"en":      "Form",
		"zh-Hans": "表单详情",
		"zh-Hant": "表單詳情",
	},
	"fano_table_tabs_fullscreen": {
		"en":      "Full Screen",
		"zh-Hans": "全屏",
		"zh-Hant": "全屏",
	},
	"fano_table_total": {
		"en":      "Total {{total}} items",
		"zh-Hans": "{{total}} 条记录",
		"zh-Hant": "{{total}} 條記錄",
	},
	"fano_icon_choose": {
		"en":      "Choose Icon",
		"zh-Hans": "选择图标",
		"zh-Hant": "選擇圖標",
	},
	"fano_icon_search_placeholder": {
		"en":      "Search icons here",
		"zh-Hans": "输入关键字搜索图标",
		"zh-Hant": "輸入關鍵字搜索圖標",
	},
	"fano_icon_theme_outlined": {
		"en":      "Outlined",
		"zh-Hans": "线框风格",
		"zh-Hant": "線框風格",
	},
	"fano_icon_theme_filled": {
		"en":      "Filled",
		"zh-Hans": "实底风格",
		"zh-Hant": "實底風格",
	},
	"fano_icon_theme_twotone": {
		"en":      "Two Tone",
		"zh-Hans": "双色风格",
		"zh-Hant": "雙色風格",
	},
	"fano_group_add_item": {
		"en":      "Add Item",
		"zh-Hans": "增加",
		"zh-Hant": "增加",
	},
	"fano_upload_button": {
		"en":      "Click to Upload",
		"zh-Hans": "点击上传",
		"zh-Hant": "點擊上傳",
	},
	"import_empty": {
		"en":      "Import data is empty",
		"zh-Hans": "导入记录为空",
		"zh-Hant": "導入記錄為空",
	},
	"import_failed": {
		"en":      "Import failed",
		"zh-Hans": "导入失败",
		"zh-Hant": "導入失敗",
	},
	"import_success": {
		"en":      "Imported successfully",
		"zh-Hans": "导入成功",
		"zh-Hant": "導入成功",
	},
	"kuu_apikeys_createdat": {
		"en":      "Created At",
		"zh-Hans": "创建时间",
		"zh-Hant": "創建時間",
	},
	"kuu_apikeys_desc": {
		"en":      "Description",
		"zh-Hans": "描述",
		"zh-Hant": "描述",
	},
	"kuu_apikeys_desc_placeholder": {
		"en":      "Optional: e.g. This key is used by the cron service to trigger jobs",
		"zh-Hans": "任务调度中心使用此密钥来触发定时任务",
		"zh-Hant": "任務調度中心使用此密鑰來觸發定時任務",
	},
	"kuu_apikeys_desc_render": {
		"en":      "User login",
		"zh-Hans": "用户登录",
		"zh-Hant": "用戶登錄",
	},
	"kuu_apikeys_desc_required": {
		"en":      "Please enter a description",
		"zh-Hans": "请输入描述",
		"zh-Hant": "請輸入描述",
	},
	"kuu_apikeys_exp": {
		"en":      "Exp",
		"zh-Hans": "过期时间",
		"zh-Hant": "過期時間",
	},
	"kuu_apikeys_exp_never_exp": {
		"en":      "Never Expire",
		"zh-Hans": "永不过期",
		"zh-Hant": "永不過期",
	},
	"kuu_apikeys_exp_options_day": {
		"en":      "A day from now",
		"zh-Hans": "从现在开始，有效期1天",
		"zh-Hant": "從現在開始，有效期1天",
	},
	"kuu_apikeys_exp_options_month": {
		"en":      "A month from now",
		"zh-Hans": "从现在开始，有效期1个月",
		"zh-Hant": "從現在開始，有效期1個月",
	},
	"kuu_apikeys_exp_options_never": {
		"en":      "Never",
		"zh-Hans": "永不过期",
		"zh-Hant": "永不過期",
	},
	"kuu_apikeys_exp_options_week": {
		"en":      "A week from now",
		"zh-Hans": "从现在开始，有效期1周",
		"zh-Hant": "從現在開始，有效期1周",
	},
	"kuu_apikeys_exp_options_year": {
		"en":      "A year from now",
		"zh-Hans": "从现在开始，有效期1年",
		"zh-Hant": "從現在開始，有效期1年",
	},
	"kuu_apikeys_exp_required": {
		"en":      "Please select automatic expiration time",
		"zh-Hans": "请选择自动过期时间",
		"zh-Hant": "請選擇自動過期時間",
	},
	"kuu_apikeys_state": {
		"en":      "State",
		"zh-Hans": "有效状态",
		"zh-Hant": "有效狀態",
	},
	"kuu_apikeys_token_copy_copied": {
		"en":      "The token has been copied",
		"zh-Hans": "令牌已复制",
		"zh-Hant": "令牌已複製",
	},
	"kuu_apikeys_token_copy_tooltip": {
		"en":      "Click to copy token",
		"zh-Hans": "点击复制令牌",
		"zh-Hant": "點擊複製令牌",
	},
	"kuu_apikeys_token_copy_txt": {
		"en":      "Copy",
		"zh-Hans": "复制令牌",
		"zh-Hant": "複製令牌",
	},
	"kuu_apikeys_token_enable_confirm": {
		"en":      "Are you sure to re-enable the token?",
		"zh-Hans": "确定重新激活该令牌吗？",
		"zh-Hant": "確定重新激活該令牌嗎？",
	},
	"kuu_apikeys_token_enable_tooltip": {
		"en":      "Enable now",
		"zh-Hans": "立即启用",
		"zh-Hant": "立即啟用",
	},
	"kuu_apikeys_token_enable_txt": {
		"en":      "Enable",
		"zh-Hans": "重新启用",
		"zh-Hant": "從新啟用",
	},
	"kuu_apikeys_token_exp_confirm": {
		"en":      "Are you sure to expire this token?",
		"zh-Hans": "确定作废此令牌吗？",
		"zh-Hant": "確定作廢此令牌嗎？",
	},
	"kuu_apikeys_token_exp_tooltip": {
		"en":      "Expired now",
		"zh-Hans": "立即过期",
		"zh-Hant": "立即過期",
	},
	"kuu_apikeys_token_exp_txt": {
		"en":      "Expire",
		"zh-Hans": "立即过期",
		"zh-Hant": "立即過期",
	},
	"kuu_common_org": {
		"en":      "Organization",
		"zh-Hans": "所属组织",
		"zh-Hant": "所屬組織",
	},
	"kuu_exception_403": {
		"en":      "Sorry, you don't have access to this page.",
		"zh-Hans": "抱歉，您无权访问此页面。",
		"zh-Hant": "抱歉，您無權訪問此頁面。",
	},
	"kuu_exception_404": {
		"en":      "Sorry, the page you visited does not exist.",
		"zh-Hans": "抱歉，您访问的页面不存在。",
		"zh-Hant": "抱歉，您訪問的頁面不存在。",
	},
	"kuu_exception_500": {
		"en":      "Sorry, the server is reporting an error.",
		"zh-Hans": "抱歉，服务器报告错误。",
		"zh-Hant": "抱歉，服務器報告錯誤。",
	},
	"kuu_exception_btn_back": {
		"en":      "back to home",
		"zh-Hans": "回到主页",
		"zh-Hant": "回到主頁",
	},
	"kuu_file_actions_upload": {
		"en":      "Upload",
		"zh-Hans": "上传文件",
		"zh-Hant": "上傳文件",
	},
	"kuu_file_class": {
		"en":      "Class",
		"zh-Hans": "文件分类",
		"zh-Hant": "文件分類",
	},
	"kuu_file_createdat": {
		"en":      "Created At",
		"zh-Hans": "上传时间",
		"zh-Hant": "上傳時間",
	},
	"kuu_file_name": {
		"en":      "Name",
		"zh-Hans": "文件名称",
		"zh-Hant": "文件名稱",
	},
	"kuu_file_preview": {
		"en":      "Preview",
		"zh-Hans": "文件预览",
		"zh-Hant": "文件預覽",
	},
	"kuu_file_size": {
		"en":      "Size",
		"zh-Hans": "文件大小",
		"zh-Hant": "文件大小",
	},
	"kuu_file_type": {
		"en":      "Type",
		"zh-Hans": "文件类型",
		"zh-Hant": "文件類型",
	},
	"kuu_file_uid": {
		"en":      "UID",
		"zh-Hans": "文件唯一ID",
		"zh-Hant": "文件唯一ID",
	},
	"kuu_file_url": {
		"en":      "URL",
		"zh-Hans": "文件地址",
		"zh-Hant": "文件地址",
	},
	"kuu_i18n_key": {
		"en":      "Key",
		"zh-Hans": "翻译键",
		"zh-Hant": "翻譯鍵",
	},
	"kuu_i18n_key_required": {
		"en":      "Key is required",
		"zh-Hans": "翻译键是必须的",
		"zh-Hant": "翻譯鍵是必須的",
	},
	"kuu_i18n_description": {
		"en":      "Description",
		"zh-Hans": "用途描述",
		"zh-Hant": "用途描述",
	},
	"kuu_i18n_upload": {
		"en":      "Click to Upload",
		"zh-Hans": "点击上传",
		"zh-Hant": "點擊上傳",
	},
	"kuu_i18n_overwrite_update": {
		"en":      "Overwrite Update",
		"zh-Hans": "覆盖更新",
		"zh-Hant": "覆蓋更新",
	},
	"kuu_i18n_incremental_update": {
		"en":      "Incremental Update",
		"zh-Hans": "增量更新",
		"zh-Hant": "增量更新",
	},
	"kuu_import_channel": {
		"en":      "Channel",
		"zh-Hans": "导入渠道",
		"zh-Hant": "導入渠道",
	},
	"kuu_import_createdat": {
		"en":      "Import Time",
		"zh-Hans": "导入时间",
		"zh-Hant": "導入時間",
	},
	"kuu_import_data": {
		"en":      "View Imported",
		"zh-Hans": "导入预览",
		"zh-Hant": "導入預覽",
	},
	"kuu_import_download": {
		"en":      "Download",
		"zh-Hans": "下载",
		"zh-Hant": "下載",
	},
	"kuu_import_feedback": {
		"en":      "View Feedback",
		"zh-Hans": "反馈预览",
		"zh-Hant": "反饋預覽",
	},
	"kuu_import_importsn": {
		"en":      "Serial No",
		"zh-Hans": "序列号",
		"zh-Hant": "序列號",
	},
	"kuu_import_message": {
		"en":      "Message",
		"zh-Hans": "导入提示",
		"zh-Hant": "導入提示",
	},
	"kuu_import_reimport": {
		"en":      "Reimport",
		"zh-Hans": "重新导入",
		"zh-Hant": "重新導入",
	},
	"kuu_import_status": {
		"en":      "Status",
		"zh-Hans": "导入状态",
		"zh-Hant": "導入狀態",
	},
	"kuu_import_status_failed": {
		"en":      "Failed",
		"zh-Hans": "导入失败",
		"zh-Hant": "導入失敗",
	},
	"kuu_import_status_importing": {
		"en":      "Importing",
		"zh-Hans": "导入中",
		"zh-Hant": "導入中",
	},
	"kuu_import_status_success": {
		"en":      "Success",
		"zh-Hans": "导入成功",
		"zh-Hant": "導入成功",
	},
	"kuu_layout_tabs_close_left": {
		"en":      "Close All to the Left",
		"zh-Hans": "关闭左侧",
		"zh-Hant": "關閉左側",
	},
	"kuu_layout_tabs_close_others": {
		"en":      "Close Others",
		"zh-Hans": "关闭其他",
		"zh-Hant": "關閉其他",
	},
	"kuu_layout_tabs_close_right": {
		"en":      "Close All to the Right",
		"zh-Hans": "关闭右侧",
		"zh-Hant": "關閉右側",
	},
	"kuu_layout_tabs_refresh": {
		"en":      "Refresh",
		"zh-Hans": "刷新",
		"zh-Hant": "刷新",
	},
	"kuu_login_btn_submit": {
		"en":      "Login",
		"zh-Hans": "登录",
		"zh-Hant": "登錄",
	},
	"kuu_login_captcha_placeholder": {
		"en":      "Captcha",
		"zh-Hans": "验证码",
		"zh-Hant": "驗證碼",
	},
	"kuu_login_captcha_required": {
		"en":      "Please enter the captcha",
		"zh-Hans": "请输入验证码",
		"zh-Hant": "請輸入驗證碼",
	},
	"kuu_login_password_forgot": {
		"en":      "Forgot your password?",
		"zh-Hans": "忘记密码？",
		"zh-Hant": "忘記密碼？",
	},
	"kuu_login_password_placeholder": {
		"en":      "Password",
		"zh-Hans": "密码",
		"zh-Hant": "密碼",
	},
	"kuu_login_password_required": {
		"en":      "Please enter your password",
		"zh-Hans": "请输入登录密码",
		"zh-Hant": "請輸入登錄密碼",
	},
	"kuu_login_remember": {
		"en":      "Remember",
		"zh-Hans": "记住我",
		"zh-Hant": "記住我",
	},
	"kuu_login_username_placeholder": {
		"en":      "Username",
		"zh-Hans": "账号",
		"zh-Hant": "賬號",
	},
	"kuu_login_username_required": {
		"en":      "Please enter your username",
		"zh-Hans": "请输入登录账号",
		"zh-Hant": "請輸入登錄賬號",
	},
	"kuu_menu_add_submenu": {
		"en":      "Add Submenu",
		"zh-Hans": "添加子级",
		"zh-Hant": "添加子級",
	},
	"kuu_menu_closeable": {
		"en":      "Closeable",
		"zh-Hans": "可关闭",
		"zh-Hant": "可關閉",
	},
	"kuu_menu_code": {
		"en":      "Permission Code",
		"zh-Hans": "权限编码",
		"zh-Hant": "權限編碼",
	},
	"kuu_menu_defaultopen": {
		"en":      "Open by default",
		"zh-Hans": "默认打开",
		"zh-Hant": "默認打開",
	},
	"kuu_menu_detail": {
		"en":      "Detail",
		"zh-Hans": "详情",
		"zh-Hant": "詳情",
	},
	"kuu_menu_disable": {
		"en":      "Disable",
		"zh-Hans": "是否禁用",
		"zh-Hant": "是否禁用",
	},
	"kuu_menu_external": {
		"en":      "External link",
		"zh-Hans": "外部链接",
		"zh-Hant": "外部鏈接",
	},
	"kuu_menu_icon": {
		"en":      "Icon",
		"zh-Hans": "图标",
		"zh-Hant": "圖標",
	},
	"kuu_menu_localekey": {
		"en":      "Locale Key",
		"zh-Hans": "国际化键",
		"zh-Hant": "國際化鍵",
	},
	"kuu_menu_name": {
		"en":      "Name",
		"zh-Hans": "菜单名称",
		"zh-Hant": "菜單名稱",
	},
	"kuu_menu_name_required": {
		"en":      "Please enter a menu name",
		"zh-Hans": "请输入菜单名称",
		"zh-Hant": "請輸入菜單名稱",
	},
	"kuu_menu_parent": {
		"en":      "Parent Menu",
		"zh-Hans": "父级菜单",
		"zh-Hant": "父級菜單",
	},
	"kuu_menu_sort": {
		"en":      "Sort",
		"zh-Hans": "菜单排序",
		"zh-Hant": "菜單排序",
	},
	"kuu_menu_uri": {
		"en":      "URI",
		"zh-Hans": "菜单地址",
		"zh-Hant": "菜單地址",
	},
	"kuu_meta_code": {
		"en":      "Module",
		"zh-Hans": "所属模块",
		"zh-Hant": "所屬模塊",
	},
	"kuu_meta_displayname": {
		"en":      "Display Name",
		"zh-Hans": "显示名",
		"zh-Hant": "顯示名",
	},
	"kuu_meta_fields_code": {
		"en":      "Field Code",
		"zh-Hans": "字段编码",
		"zh-Hant": "字段編碼",
	},
	"kuu_meta_fields_enum": {
		"en":      "Enum",
		"zh-Hans": "关联枚举",
		"zh-Hant": "關聯枚舉",
	},
	"kuu_meta_fields_isarray": {
		"en":      "Is Array",
		"zh-Hans": "是否数组",
		"zh-Hant": "是否數組",
	},
	"kuu_meta_fields_ispassword": {
		"en":      "Is Password",
		"zh-Hans": "是否密码",
		"zh-Hant": "是否密碼",
	},
	"kuu_meta_fields_isref": {
		"en":      "Is Ref",
		"zh-Hans": "是否引用",
		"zh-Hant": "是否引用",
	},
	"kuu_meta_fields_kind": {
		"en":      "Kind",
		"zh-Hans": "字段类型",
		"zh-Hant": "字段類型",
	},
	"kuu_meta_fields_name": {
		"en":      "Field Name",
		"zh-Hans": "字段名称",
		"zh-Hant": "字段名稱",
	},
	"kuu_meta_name": {
		"en":      "Name",
		"zh-Hans": "名称",
		"zh-Hant": "名稱",
	},
	"kuu_navbar_apiendpoint": {
		"en":      "API Endpoint",
		"zh-Hans": "API服务器",
		"zh-Hant": "API服務器",
	},
	"kuu_navbar_apiendpoint_placeholder": {
		"en":      "Optional: e.g. https://kuu.example.com/api",
		"zh-Hans": "例如：https://kuu.example.com/api",
		"zh-Hant": "例如：https://kuu.example.com/api",
	},
	"kuu_navbar_apikeys": {
		"en":      "API Keys",
		"zh-Hans": "API Keys",
		"zh-Hant": "API Keys",
	},
	"kuu_navbar_changepass": {
		"en":      "Change password",
		"zh-Hans": "修改密码",
		"zh-Hant": "修改密碼",
	},
	"kuu_navbar_languages": {
		"en":      "Languages",
		"zh-Hans": "语言切换",
		"zh-Hant": "語言切換",
	},
	"kuu_navbar_loginas": {
		"en":      "Login As",
		"zh-Hans": "模拟登录",
		"zh-Hant": "模拟登录",
	},
	"kuu_navbar_loginas_placeholder": {
		"en":      "Select a user",
		"zh-Hans": "选择一个用户",
		"zh-Hant": "選擇一個用戶",
	},
	"kuu_navbar_logout": {
		"en":      "Logout",
		"zh-Hans": "退出登录",
		"zh-Hant": "退出登錄",
	},
	"kuu_navbar_profile": {
		"en":      "Profile",
		"zh-Hans": "个人中心",
		"zh-Hant": "個人中心",
	},
	"kuu_org_btn_login": {
		"en":      "Login",
		"zh-Hans": "登录",
		"zh-Hant": "登錄",
	},
	"kuu_org_code": {
		"en":      "Code",
		"zh-Hans": "组织编码",
		"zh-Hant": "組織編碼",
	},
	"kuu_org_createdat": {
		"en":      "Created At",
		"zh-Hans": "创建时间",
		"zh-Hant": "創建時間",
	},
	"kuu_org_name": {
		"en":      "Name",
		"zh-Hans": "组织名称",
		"zh-Hant": "組織名稱",
	},
	"kuu_org_parent": {
		"en":      "Parent",
		"zh-Hans": "上级组织",
		"zh-Hant": "上級組織",
	},
	"kuu_org_select_login": {
		"en":      "Please select a login organization",
		"zh-Hans": "请选择登录组织",
		"zh-Hant": "請選擇登入組織",
	},
	"kuu_org_sort": {
		"en":      "Sort",
		"zh-Hans": "排序",
		"zh-Hant": "排序",
	},
	"kuu_org_unorganized": {
		"en":      "You have not assigned any organization",
		"zh-Hans": "您尚未分配任何组织",
		"zh-Hant": "您尚未分配任何組織",
	},
	"kuu_param_builtin": {
		"en":      "Built-in",
		"zh-Hans": "是否内置",
		"zh-Hant": "是否內置",
	},
	"kuu_param_code": {
		"en":      "Code",
		"zh-Hans": "参数编码",
		"zh-Hant": "參數編碼",
	},
	"kuu_param_createdat": {
		"en":      "Created At",
		"zh-Hans": "创建时间",
		"zh-Hant": "創建時間",
	},
	"kuu_param_name": {
		"en":      "Name",
		"zh-Hans": "参数名称",
		"zh-Hant": "參數名稱",
	},
	"kuu_param_sort": {
		"en":      "Sort",
		"zh-Hans": "排序",
		"zh-Hant": "排序",
	},
	"kuu_param_type": {
		"en":      "Type",
		"zh-Hans": "值类型",
		"zh-Hant": "值類型",
	},
	"kuu_param_type_color": {
		"en":      "Color",
		"zh-Hans": "顏色组件",
		"zh-Hant": "顏色組件",
	},
	"kuu_param_type_datepicker": {
		"en":      "DatePicker",
		"zh-Hans": "日期选择框",
		"zh-Hant": "日期選擇框",
	},
	"kuu_param_type_editor": {
		"en":      "Editor",
		"zh-Hans": "编辑器",
		"zh-Hant": "編輯器",
	},
	"kuu_param_type_icon": {
		"en":      "Icon",
		"zh-Hans": "图标组件",
		"zh-Hant": "圖標組件",
	},
	"kuu_param_type_input": {
		"en":      "Input",
		"zh-Hans": "输入框",
		"zh-Hant": "輸入框",
	},
	"kuu_param_type_json": {
		"en":      "JSON",
		"zh-Hans": "JSON编辑器",
		"zh-Hant": "JSON編輯器",
	},
	"kuu_param_type_monthpicker": {
		"en":      "MonthPicker",
		"zh-Hans": "月份选择框",
		"zh-Hant": "月份選擇框",
	},
	"kuu_param_type_number": {
		"en":      "Number",
		"zh-Hans": "数字输入框",
		"zh-Hant": "數字輸入框",
	},
	"kuu_param_type_password": {
		"en":      "Password",
		"zh-Hans": "密码输入框",
		"zh-Hant": "密碼輸入框",
	},
	"kuu_param_type_rangepicker": {
		"en":      "RangePicker",
		"zh-Hans": "范围选择框",
		"zh-Hant": "範圍選擇框",
	},
	"kuu_param_type_switch": {
		"en":      "Switch",
		"zh-Hans": "开关",
		"zh-Hant": "開關",
	},
	"kuu_param_type_textarea": {
		"en":      "Textarea",
		"zh-Hans": "多行文本",
		"zh-Hant": "多行文本",
	},
	"kuu_param_type_timepicker": {
		"en":      "TimePicker",
		"zh-Hans": "时间选择框",
		"zh-Hant": "時間選擇框",
	},
	"kuu_param_type_upload": {
		"en":      "Upload",
		"zh-Hans": "上传组件",
		"zh-Hant": "上傳組件",
	},
	"kuu_param_type_weekpicker": {
		"en":      "WeekPicker",
		"zh-Hans": "周选择框",
		"zh-Hant": "周選擇框",
	},
	"kuu_param_value": {
		"en":      "Value",
		"zh-Hans": "参数值",
		"zh-Hant": "參數值",
	},
	"kuu_permission_add_sub": {
		"en":      "Add Sub",
		"zh-Hans": "添加子级",
		"zh-Hant": "添加子級",
	},
	"kuu_permission_code": {
		"en":      "Permission Code",
		"zh-Hans": "权限编码",
		"zh-Hant": "權限編碼",
	},
	"kuu_permission_disable": {
		"en":      "Disable",
		"zh-Hans": "是否禁用",
		"zh-Hant": "是否禁用",
	},
	"kuu_permission_name": {
		"en":      "Name",
		"zh-Hans": "权限名称",
		"zh-Hant": "權限名稱",
	},
	"kuu_permission_name_required": {
		"en":      "Please enter a permission name",
		"zh-Hans": "请输入权限名称",
		"zh-Hant": "請輸入權限名稱",
	},
	"kuu_permission_parent": {
		"en":      "Parent",
		"zh-Hans": "父级权限",
		"zh-Hant": "父級權限",
	},
	"kuu_permission_sort": {
		"en":      "Sort",
		"zh-Hans": "权限排序",
		"zh-Hant": "權限排序",
	},
	"kuu_role_addrule": {
		"en":      "Add rule",
		"zh-Hans": "添加规则",
		"zh-Hant": "添加規則",
	},
	"kuu_role_builtin": {
		"en":      "Built-in",
		"zh-Hans": "是否内置",
		"zh-Hant": "是否內置",
	},
	"kuu_role_code": {
		"en":      "Code",
		"zh-Hans": "角色编码",
		"zh-Hant": "角色編碼",
	},
	"kuu_role_code_required": {
		"en":      "Please enter a role code",
		"zh-Hans": "请输入角色编码",
		"zh-Hant": "請輸入角色編碼",
	},
	"kuu_role_createdat": {
		"en":      "Created At",
		"zh-Hans": "创建时间",
		"zh-Hant": "創建時間",
	},
	"kuu_role_data_range_current": {
		"en":      "CURRENT",
		"zh-Hans": "当前组织",
		"zh-Hant": "當前組織",
	},
	"kuu_role_data_range_current_following": {
		"en":      "CURRENT_FOLLOWING",
		"zh-Hans": "当前组织及以下",
		"zh-Hant": "當前組織及以下",
	},
	"kuu_role_data_range_personal": {
		"en":      "PERSONAL",
		"zh-Hans": "个人范围",
		"zh-Hant": "個人範圍",
	},
	"kuu_role_dp": {
		"en":      "Data Privileges",
		"zh-Hans": "数据权限",
		"zh-Hant": "數據權限",
	},
	"kuu_role_name": {
		"en":      "Name",
		"zh-Hans": "角色名称",
		"zh-Hant": "角色名稱",
	},
	"kuu_role_name_required": {
		"en":      "Please enter a role name",
		"zh-Hans": "请输入角色名称",
		"zh-Hant": "請輸入角色名稱",
	},
	"kuu_role_op": {
		"en":      "Menu Privileges",
		"zh-Hans": "菜单权限",
		"zh-Hant": "菜單權限",
	},
	"kuu_role_readable_range_placeholder": {
		"en":      "Please select a readable range",
		"zh-Hans": "请选择可读数据范围",
		"zh-Hant": "請選擇可讀數據範圍",
	},
	"kuu_role_select_org_placeholder": {
		"en":      "Please select an organization",
		"zh-Hans": "请选择组织",
		"zh-Hant": "請選擇組織",
	},
	"kuu_role_writable_range_placeholder": {
		"en":      "Please select a writable range",
		"zh-Hans": "请选择可写数据范围",
		"zh-Hant": "請選擇可寫數據範圍",
	},
	"kuu_up": {
		"en":      "{{time}}",
		"zh-Hans": "{{time}}",
		"zh-Hant": "{{time}}",
	},
	"kuu_user_builtin": {
		"en":      "Built-in",
		"zh-Hans": "是否内置",
		"zh-Hant": "是否內置",
	},
	"kuu_user_createdat": {
		"en":      "Created At",
		"zh-Hans": "创建时间",
		"zh-Hant": "創建時間",
	},
	"kuu_user_disable": {
		"en":      "Disable",
		"zh-Hans": "是否禁用",
		"zh-Hant": "是否禁用",
	},
	"kuu_user_name": {
		"en":      "Real name",
		"zh-Hans": "真实姓名",
		"zh-Hant": "真實姓名",
	},
	"kuu_user_password": {
		"en":      "Password",
		"zh-Hans": "密码",
		"zh-Hant": "密碼",
	},
	"kuu_user_role_assigns": {
		"en":      "Role Assignments",
		"zh-Hans": "角色分配",
		"zh-Hant": "角色分配",
	},
	"kuu_user_titles_assigned": {
		"en":      "Assigned",
		"zh-Hans": "已分配角色",
		"zh-Hant": "已分配角色",
	},
	"kuu_user_titles_notassigned": {
		"en":      "Not Assigned",
		"zh-Hans": "未分配角色",
		"zh-Hant": "未分配角色",
	},
	"kuu_user_username": {
		"en":      "Username",
		"zh-Hans": "用户名",
		"zh-Hant": "用戶名",
	},
	"lang_list_save_failed": {
		"en":      "Save languages failed",
		"zh-Hans": "保存语言列表失败",
		"zh-Hant": "保存語言列表失敗",
	},
	"lang_msgs_failed": {
		"en":      "Query messages failed",
		"zh-Hans": "查询国际化配置失败",
		"zh-Hant": "查詢國際化配置失敗",
	},
	"lang_switch_failed": {
		"en":      "Language switching failed",
		"zh-Hans": "语言切换失败",
		"zh-Hant": "語言切換失敗",
	},
	"lang_trans_query_failed": {
		"en":      "Query translation list failed",
		"zh-Hans": "查询国际化翻译列表失败",
		"zh-Hant": "查詢國際化翻譯列表失敗",
	},
	"lang_trans_save_failed": {
		"en":      "Save locale messages failed",
		"zh-Hans": "保存国际化配置失败",
		"zh-Hant": "保存國際化配置失敗",
	},
	"login_as_failed": {
		"en":      "Login failed",
		"zh-Hans": "模拟登录失败",
		"zh-Hant": "模擬登錄失敗",
	},
	"login_as_unauthorized": {
		"en":      "Unauthorized operation",
		"zh-Hans": "无权操作",
		"zh-Hant": "無權操作",
	},
	"menu_default": {
		"en":      "Default Menu",
		"zh-Hans": "默认菜单",
		"zh-Hant": "默認菜單",
	},
	"menu_auth": {
		"en":      "Authorization Management",
		"zh-Hans": "权限管理",
		"zh-Hant": "權限管理",
	},
	"menu_auth_menu": {
		"en":      "Menu Management",
		"zh-Hans": "菜单管理",
		"zh-Hant": "菜單管理",
	},
	"menu_auth_org": {
		"en":      "Organization Management",
		"zh-Hans": "组织管理",
		"zh-Hant": "組織管理",
	},
	"menu_auth_user": {
		"en":      "User Management",
		"zh-Hans": "用户管理",
		"zh-Hant": "用戶管理",
	},
	"menu_auth_role": {
		"en":      "Role Management",
		"zh-Hans": "角色管理",
		"zh-Hant": "角色管理",
	},
	"menu_auth_permission": {
		"en":      "Permission Management",
		"zh-Hans": "权限管理",
		"zh-Hant": "權限管理",
	},
	"menu_sys": {
		"en":      "System Management",
		"zh-Hans": "系统管理",
		"zh-Hant": "系統管理",
	},
	"menu_sys_param": {
		"en":      "Param Management",
		"zh-Hans": "参数管理",
		"zh-Hant": "參數管理",
	},
	"menu_sys_file": {
		"en":      "File Management",
		"zh-Hans": "文件管理",
		"zh-Hant": "上傳管理",
	},
	"menu_sys_import": {
		"en":      "Import Management",
		"zh-Hans": "导入文件",
		"zh-Hant": "導入文件",
	},
	"menu_sys_languages": {
		"en":      "Languages",
		"zh-Hans": "多语言",
		"zh-Hant": "多語言",
	},
	"model_docs_failed": {
		"en":      "Model document query failed",
		"zh-Hans": "默认接口文档查询失败",
		"zh-Hant": "默認接口文檔查詢失敗",
	},
	"org_login_failed": {
		"en":      "Organization login failed",
		"zh-Hans": "组织登入失败",
		"zh-Hant": "組織登入失敗",
	},
	"org_not_found": {
		"en":      "Organization not found",
		"zh-Hans": "组织不存在",
		"zh-Hant": "組織不存在",
	},
	"org_query_failed": {
		"en":      "Organization query failed",
		"zh-Hans": "组织列表查询失败",
		"zh-Hant": "組織列表查詢失敗",
	},
	"reimport_failed": {
		"en":      "Reimport failed",
		"zh-Hans": "重新导入失败",
		"zh-Hant": "重新導入失敗",
	},
	"rest_create_failed": {
		"en":      "Create failed",
		"zh-Hans": "新增失败",
		"zh-Hant": "新增失敗",
	},
	"rest_delete_failed": {
		"en":      "Delete failed",
		"zh-Hans": "删除失败",
		"zh-Hant": "刪除失敗",
	},
	"rest_export_failed": {
		"en":      "Export failed",
		"zh-Hans": "导出失败",
		"zh-Hant": "導出失敗",
	},
	"rest_import_failed": {
		"en":      "Import failed",
		"zh-Hans": "导入失败",
		"zh-Hant": "導入失敗",
	},
	"rest_query_failed": {
		"en":      "Query failed",
		"zh-Hans": "查询失败",
		"zh-Hant": "查詢失敗",
	},
	"rest_update_failed": {
		"en":      "Update failed",
		"zh-Hans": "更新失败",
		"zh-Hant": "更新失敗",
	},
	"role_assigns_failed": {
		"en":      "User roles query failed",
		"zh-Hans": "用户角色查询失败",
		"zh-Hant": "用戶角色查詢失敗",
	},
	"sys_meta_failed": {
		"en":      "Metadata does not exist: {{name}}",
		"zh-Hans": "元数据不存在：{{name}}",
		"zh-Hant": "元數據不存在：{{name}}",
	},
	"upload_failed": {
		"en":      "Upload file failed",
		"zh-Hans": "文件上传失败",
		"zh-Hant": "文件上傳失敗",
	},
	"user_menus_failed": {
		"en":      "User menus query failed",
		"zh-Hans": "用户菜单查询失败",
		"zh-Hant": "用戶菜單查詢失敗",
	},
}

func AddPresetIntlMessage(messages map[string]map[string]string) {
	presetIntlMessagesMu.Lock()
	defer presetIntlMessagesMu.Unlock()

	for k, values := range messages {
		if len(values) == 0 {
			continue
		}
		if presetIntlMessages[k] == nil {
			presetIntlMessages[k] = make(map[string]string)
		}
		for l, v := range values {
			presetIntlMessages[k][l] = v
		}
	}
}

func AddDefaultIntlMessage(messages map[string]string) {
	m := make(map[string]map[string]string)
	languageMap := intl.LanguageMap()
	for k, defaultText := range messages {
		if m[k] == nil {
			m[k] = make(map[string]string)
		}
		for l := range languageMap {
			m[k][l] = defaultText
		}
	}
	AddPresetIntlMessage(m)
}

func GetPresetIntlMessage() map[string]map[string]string {
	presetIntlMessagesMu.RLock()
	defer presetIntlMessagesMu.RUnlock()

	return presetIntlMessages
}
