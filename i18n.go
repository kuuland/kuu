package kuu

const (
	ErrParsingFailed    = "ERR_PARSING_FAILED"    // -1 网络错误，参数解析失败
	ErrValidationFailed = "ERR_VALIDATION_FAILED" // -1 网络错误，参数验证失败
	ErrUnauthorized     = "ERR_UNAUTHORIZED"      // 401 身份验证失败，请刷新后重试
	ErrForbidden        = "ERR_FORBIDDEN"         // 403 抱歉，您无访问权限
	ErrInternalServer   = "ERR_INTERNAL_SERVER"   // 500 系统错误，请稍后重试
	ErrOperationFailed  = "ERR_OPERATION_FAILED"  // 100 操作失败，请稍后重试
	ErrSignInFailed     = "ERR_SIGN_IN_FAILED"    // 100 登录失败，请稍后重试
	ErrSignOutFailed    = "ERR_SIGN_OUT_FAILED"   // 101 退出失败，请稍后重试
	ErrSignUpFailed     = "ERR_SIGN_UP_FAILED"    // 102 注册失败，请稍后重试
	ErrUploadFailed     = "ERR_UPLOAD_FAILED"     // 103 上传失败，请稍后重试

	ErrCreateFailed = "ERR_CREATE_FAILED" // 104 创建失败，请稍后重试
	ErrUpdateFailed = "ERR_UPDATE_FAILED" // 105 更新失败，请稍后重试
	ErrDeleteFailed = "ERR_DELETE_FAILED" // 106 删除失败，请稍后重试
	ErrQueryFailed  = "ERR_QUERY_FAILED"  // 107 查询失败，请稍后重试
	ErrSaveFailed   = "ERR_SAVE_FAILED"   // 108 保存失败，请稍后重试
)
