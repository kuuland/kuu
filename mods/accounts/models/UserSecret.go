package models

import "time"

// UserSecret 用户密钥
type UserSecret struct {
	ID     string `json:"_id" displayName:"用户密钥"`
	UserID string `name:"关联用户ID"`
	Secret string `name:"使用密钥"`
	Token  string `name:"令牌信息"`
	// 标准字段
	CreatedAt time.Time `name:"创建时间"`
	UpdatedAt time.Time `name:"修改时间"`
	IsDeleted bool      `name:"是否已删除"`
	Remark    string    `name:"备注"`
}
