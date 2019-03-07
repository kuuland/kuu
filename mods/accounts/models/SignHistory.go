package models

// SignHistory 用户登入/登出历史
type SignHistory struct {
	ID        string      `json:"_id" displayName:"用户登入/登出历史" noauth:"true"`
	ReqData   interface{} `name:"请求数据"`
	LoginData interface{} `name:"登录数据"`
	Token     string      `name:"令牌信息"`
	Method    string      `name:"登录方式" remark:"login/logout"`
	// 标准字段
	CreatedAt int64  `name:"创建时间"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}
