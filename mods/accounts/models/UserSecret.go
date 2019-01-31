package models

// UserSecret 用户密钥
type UserSecret struct {
	ID     string `json:"_id" displayName:"用户密钥"`
	UserID string `name:"关联用户ID"`
	Secret string `name:"使用密钥"`
	Token  string `name:"令牌信息"`
	Iat    int64  `name:"签发时间戳"`
	Exp    int64  `name:"过期时间戳"`
	Method string `name:"登录方式" remark:"login/logout"`
	// 标准字段
	CreatedAt int64 `name:"创建时间"`
	UpdatedAt int64 `name:"修改时间"`
	IsDeleted bool      `name:"是否已删除"`
	Remark    string    `name:"备注"`
}
